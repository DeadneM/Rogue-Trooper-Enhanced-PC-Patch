# Security / patch integrity

The patcher never patches an unknown executable by offset guessing.

It requires the exact supported retail SHA-256:

`f35d8ae6f52cbebf0d1427de5a573d2b8d4be1fe38154b8a7d44cd7ffb6b7525`

Before installation it reconstructs the target in memory and verifies the exact V64 SHA-256:

`cb2996b07333df68456afbe992242bba29609e7ffcde05ea5ee625e428103a5d`

The staged file is hashed again on disk before replacement, and the installed file is hashed once more afterwards. If final verification fails, the patcher attempts to restore the rollback copy.

The first verified vanilla backup is preserved as `RogueTrooper.exe.Backup` (or `RogueTrooper.exe.Backup.Vanilla` if the first name is occupied by another file).
