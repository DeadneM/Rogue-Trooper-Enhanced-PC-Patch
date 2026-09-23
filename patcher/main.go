package main

import (
    "bufio"
    "bytes"
    "compress/zlib"
    "crypto/sha256"
    _ "embed"
    "encoding/binary"
    "encoding/hex"
    "errors"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "strings"
)

const (
    BuildName        = "V64"
    GameExeName      = "RogueTrooper.exe"
    RetailSHA256     = "f35d8ae6f52cbebf0d1427de5a573d2b8d4be1fe38154b8a7d44cd7ffb6b7525"
    TargetSHA256     = "cb2996b07333df68456afbe992242bba29609e7ffcde05ea5ee625e428103a5d"
    RetailSize int64 = 963584
    TargetSize int64 = 2959360
)

//go:embed rogue_v64.rtdp1.zlib
var embeddedPatch []byte

type options struct {
    gameDir string
    noPause bool
}

func parseArgs() options {
    var o options
    args := os.Args[1:]
    for i := 0; i < len(args); i++ {
        switch args[i] {
        case "--game-dir":
            if i+1 < len(args) {
                i++
                o.gameDir = args[i]
            }
        case "--no-pause":
            o.noPause = true
        }
    }
    return o
}

func pause(noPause bool) {
    if noPause {
        return
    }
    fmt.Print("\nPress Enter to close...")
    _, _ = bufio.NewReader(os.Stdin).ReadString('\n')
}

func sha256Bytes(data []byte) string {
    sum := sha256.Sum256(data)
    return hex.EncodeToString(sum[:])
}

func sha256File(path string) (string, error) {
    f, err := os.Open(path)
    if err != nil {
        return "", err
    }
    defer f.Close()
    h := sha256.New()
    if _, err := io.Copy(h, f); err != nil {
        return "", err
    }
    return hex.EncodeToString(h.Sum(nil)), nil
}

func appDir(o options) (string, error) {
    if o.gameDir != "" {
        return filepath.Abs(o.gameDir)
    }
    exe, err := os.Executable()
    if err != nil {
        return "", err
    }
    return filepath.Dir(exe), nil
}

func copyFile(src, dst string) error {
    in, err := os.Open(src)
    if err != nil {
        return err
    }
    defer in.Close()

    st, err := in.Stat()
    if err != nil {
        return err
    }

    out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, st.Mode())
    if err != nil {
        return err
    }
    ok := false
    defer func() {
        _ = out.Close()
        if !ok {
            _ = os.Remove(dst)
        }
    }()

    if _, err := io.Copy(out, in); err != nil {
        return err
    }
    if err := out.Sync(); err != nil {
        return err
    }
    if err := out.Close(); err != nil {
        return err
    }
    ok = true
    return nil
}

func backupRetail(exe string) (string, error) {
    candidates := []string{exe + ".Backup", exe + ".Backup.Vanilla"}
    for _, candidate := range candidates {
        if _, err := os.Stat(candidate); err == nil {
            h, err := sha256File(candidate)
            if err == nil && strings.EqualFold(h, RetailSHA256) {
                return candidate, nil
            }
            continue
        } else if !os.IsNotExist(err) {
            return "", err
        }

        if err := copyFile(exe, candidate); err != nil {
            return "", err
        }
        h, err := sha256File(candidate)
        if err != nil {
            return "", err
        }
        if !strings.EqualFold(h, RetailSHA256) {
            _ = os.Remove(candidate)
            return "", errors.New("backup verification failed")
        }
        return candidate, nil
    }
    return "", errors.New("existing backup names are occupied by non-retail files")
}

func applyRTDP1(source, compressed []byte) ([]byte, error) {
    zr, err := zlib.NewReader(bytes.NewReader(compressed))
    if err != nil {
        return nil, fmt.Errorf("invalid compressed patch: %w", err)
    }
    raw, err := io.ReadAll(zr)
    if closeErr := zr.Close(); err == nil {
        err = closeErr
    }
    if err != nil {
        return nil, fmt.Errorf("cannot decompress patch: %w", err)
    }

    r := bytes.NewReader(raw)
    magic := make([]byte, 6)
    if _, err := io.ReadFull(r, magic); err != nil || string(magic) != "RTDP1\x00" {
        return nil, errors.New("invalid patch format")
    }

    var sourceSize, targetSize uint64
    if err := binary.Read(r, binary.LittleEndian, &sourceSize); err != nil {
        return nil, errors.New("truncated patch header")
    }
    if err := binary.Read(r, binary.LittleEndian, &targetSize); err != nil {
        return nil, errors.New("truncated patch header")
    }

    sourceSHA := make([]byte, 32)
    targetSHA := make([]byte, 32)
    if _, err := io.ReadFull(r, sourceSHA); err != nil {
        return nil, errors.New("truncated source hash")
    }
    if _, err := io.ReadFull(r, targetSHA); err != nil {
        return nil, errors.New("truncated target hash")
    }

    var count uint32
    if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
        return nil, errors.New("truncated instruction count")
    }

    actualSource := sha256.Sum256(source)
    if uint64(len(source)) != sourceSize || !bytes.Equal(actualSource[:], sourceSHA) {
        return nil, errors.New("patch source verification failed")
    }

    out := bytes.NewBuffer(make([]byte, 0, int(targetSize)))
    for i := uint32(0); i < count; i++ {
        opcode, err := r.ReadByte()
        if err != nil {
            return nil, errors.New("truncated patch instruction stream")
        }
        switch opcode {
        case 0:
            var offset uint64
            var length uint32
            if err := binary.Read(r, binary.LittleEndian, &offset); err != nil {
                return nil, errors.New("truncated copy instruction")
            }
            if err := binary.Read(r, binary.LittleEndian, &length); err != nil {
                return nil, errors.New("truncated copy instruction")
            }
            end := offset + uint64(length)
            if end > uint64(len(source)) {
                return nil, errors.New("copy range outside source")
            }
            out.Write(source[offset:end])
        case 1:
            var length uint32
            if err := binary.Read(r, binary.LittleEndian, &length); err != nil {
                return nil, errors.New("truncated literal instruction")
            }
            literal := make([]byte, int(length))
            if _, err := io.ReadFull(r, literal); err != nil {
                return nil, errors.New("truncated literal data")
            }
            out.Write(literal)
        default:
            return nil, fmt.Errorf("unknown patch opcode %d", opcode)
        }
    }

    if r.Len() != 0 {
        return nil, errors.New("unexpected trailing patch data")
    }

    result := out.Bytes()
    actualTarget := sha256.Sum256(result)
    if uint64(len(result)) != targetSize || !bytes.Equal(actualTarget[:], targetSHA) {
        return nil, errors.New("patch target verification failed")
    }
    return result, nil
}

func installVerified(exe string, target []byte) error {
    dir := filepath.Dir(exe)
    stage := filepath.Join(dir, GameExeName+".PatchNew")
    rollback := filepath.Join(dir, GameExeName+".RollbackTemp")
    _ = os.Remove(stage)
    _ = os.Remove(rollback)

    if err := os.WriteFile(stage, target, 0755); err != nil {
        return err
    }
    h, err := sha256File(stage)
    if err != nil {
        _ = os.Remove(stage)
        return err
    }
    if !strings.EqualFold(h, TargetSHA256) {
        _ = os.Remove(stage)
        return errors.New("staged V64 verification failed")
    }

    if err := os.Rename(exe, rollback); err != nil {
        _ = os.Remove(stage)
        return fmt.Errorf("cannot prepare rollback file: %w", err)
    }

    if err := os.Rename(stage, exe); err != nil {
        _ = os.Rename(rollback, exe)
        _ = os.Remove(stage)
        return fmt.Errorf("cannot install patched executable: %w", err)
    }

    finalHash, err := sha256File(exe)
    if err != nil || !strings.EqualFold(finalHash, TargetSHA256) {
        _ = os.Remove(exe)
        _ = os.Rename(rollback, exe)
        if err != nil {
            return fmt.Errorf("final verification failed: %w", err)
        }
        return errors.New("final V64 hash mismatch; original restored")
    }

    _ = os.Remove(rollback)
    return nil
}

func main() {
    o := parseArgs()
    exitCode := 0
    defer func() {
        pause(o.noPause)
        os.Exit(exitCode)
    }()

    fmt.Println(strings.Repeat("=", 68))
    fmt.Println(" Rogue Trooper - Enhanced PC Patch")
    fmt.Printf(" Stable cumulative build: %s | Steam\n", BuildName)
    fmt.Println(strings.Repeat("=", 68))

    root, err := appDir(o)
    if err != nil {
        fmt.Println("[ERROR] Cannot determine patcher directory:", err)
        exitCode = 1
        return
    }
    exe := filepath.Join(root, GameExeName)
    fmt.Println("[OK] Game directory:", root)

    if st, err := os.Stat(exe); err != nil || st.IsDir() {
        fmt.Printf("[ERROR] %s not found next to the patcher.\n", GameExeName)
        fmt.Println("        No files have been modified.")
        exitCode = 2
        return
    }

    fmt.Println("[ .. ] Verifying game executable...")
    currentHash, err := sha256File(exe)
    if err != nil {
        fmt.Println("[ERROR] Cannot hash game executable:", err)
        exitCode = 1
        return
    }

    if strings.EqualFold(currentHash, TargetSHA256) {
        fmt.Printf("[OK] %s is already installed. Nothing to do.\n", BuildName)
        return
    }
    if !strings.EqualFold(currentHash, RetailSHA256) {
        fmt.Println("[ERROR] Unsupported RogueTrooper.exe.")
        fmt.Println("        No files have been modified.")
        fmt.Println("        SHA-256:", currentHash)
        exitCode = 2
        return
    }

    st, _ := os.Stat(exe)
    if st.Size() != RetailSize {
        fmt.Println("[ERROR] Retail size check failed. No files have been modified.")
        exitCode = 2
        return
    }

    fmt.Println("[OK] Original Steam executable detected.")
    backup, err := backupRetail(exe)
    if err != nil {
        fmt.Println("[ERROR] Could not create a verified vanilla backup:", err)
        exitCode = 1
        return
    }
    fmt.Println("[OK] Backup:", filepath.Base(backup))

    fmt.Printf("[ .. ] Building cumulative %s...\n", BuildName)
    source, err := os.ReadFile(exe)
    if err != nil {
        fmt.Println("[ERROR] Cannot read source executable:", err)
        exitCode = 1
        return
    }
    target, err := applyRTDP1(source, embeddedPatch)
    if err != nil {
        fmt.Println("[ERROR]", err)
        fmt.Println("No unverified patched output was installed.")
        exitCode = 1
        return
    }
    if int64(len(target)) != TargetSize || !strings.EqualFold(sha256Bytes(target), TargetSHA256) {
        fmt.Println("[ERROR] Rebuilt V64 failed final in-memory verification.")
        fmt.Println("No unverified patched output was installed.")
        exitCode = 1
        return
    }

    fmt.Println("[ .. ] Verifying patched executable before installation...")
    fmt.Println("[OK] V64 target hash verified.")

    if err := installVerified(exe, target); err != nil {
        fmt.Println("[ERROR]", err)
        fmt.Println("No unverified patched output was intentionally left installed.")
        exitCode = 1
        return
    }

    fmt.Println(strings.Repeat("=", 68))
    fmt.Println("[SUCCESS] Rogue Trooper Enhanced PC Patch V64 installed.")
    fmt.Println("          Final SHA-256 verification passed.")
    fmt.Println(strings.Repeat("=", 68))
}
