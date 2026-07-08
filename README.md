# Kompressor

Kompressor is a small utility for cleaning and normalizing text (default `txt`) files inside a directory.

It is designed mainly for blocklist / allowlist style text files.

## Key purposes

- Remove empty lines from text files.
- Remove comment lines from text files.
- Sort lines alphabetically.
- Remove duplicate lines inside each file.
- Optionally remove duplicated files with identical content.

## Important behavior

By default, Kompressor **does not remove files**.

It only processes every file in the target directory and rewrites each file with:

- empty lines removed
- comment lines removed
- duplicate lines removed
- lines sorted

Duplicate file removal is disabled by default and must be explicitly enabled with the `-remove` flag.

This is important when files are used as public endpoints, for example:

```text
/output/whitelist.txt
/output/myzbldallow.txt
/output/myzbldblock.txt
```

## Usage

```bash
./kompressor /path/to/folder/
```

Remove duplicated files:

```bash
./kompressor /path/to/folder/ -remove
```

Alternative path syntax:

```bash
./kompressor -path /path/to/folder/
```

Example with duplicate file removal:

```bash
./kompressor -path /opt/benZine/output -remove
```

Process only `log` and `ini` files:

```bash
kompressor ./my_dir -ext log,ini
```

Exclude specific files `bld` and `csv`:

```bash
kompressor ./my_dir -exclude bld,csv
```

Allow processing `txt` and `cfg` files, but strictly block `bld`:

```bash
kompressor ./my_dir -ext txt,cfg -exclude bld
```

## Code integration

```go
cmd := exec.Command(kompressorPath, cleanPath)
```

```go
cmd := exec.Command(kompressorPath, cleanPath, "-remove")
```


