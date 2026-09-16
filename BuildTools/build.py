"""Build inside Linux/Docker with pinned Go modules and embedded Fyne metadata.

Linux tests execute under Xvfb. Windows tests are cross-compiled, not executed.
The normal Windows `fyne package --release` workflow remains supported.
"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import tomllib
import zipfile


def snapshot(source, destination):
    names = subprocess.check_output([
        "git", "-c", f"safe.directory={source}", "-C", str(source), "ls-files", "-z",
    ]).decode().split("\0")
    # Build the current worktree, including new Go source and embedded resources
    # before commit. Ignore generated builds, profiles and unrelated artifacts.
    untracked = subprocess.check_output([
        "git", "-c", f"safe.directory={source}", "-C", str(source),
        "ls-files", "--others", "--exclude-standard", "-z",
    ]).decode().split("\0")
    names += [name for name in untracked if name.endswith(".go") or name.startswith(("Resources/", "BuildTools/"))]
    for name in sorted(set(names) - {""}):
        original = source / name
        if original.is_file():
            target = destination / name
            target.parent.mkdir(parents=True, exist_ok=True)
            shutil.copy2(original, target)


def run(command, source, environment):
    print("+", " ".join(map(str, command)), flush=True)
    subprocess.run(command, cwd=source, env=environment, check=True)


def build(source, args):
    if args.test:
        formatting = subprocess.check_output(["gofmt", "-l", *map(str, sorted(source.rglob("*.go")))], text=True)
        if formatting:
            raise ValueError(f"Go source needs formatting:\n{formatting}")
    config = tomllib.loads((source / "FyneApp.toml").read_text(encoding="utf-8"))
    details = config["Details"]
    if not details.get("ID") or not details.get("Version") or not isinstance(details.get("Build"), int):
        raise ValueError("FyneApp.toml must contain ID, Version and an integer Build")
    version = f"{details['Version']}.{details['Build']}"
    release = args.release or (args.release_if_tag and os.environ.get("DRONE_BUILD_EVENT") == "tag")
    if release and os.environ.get("DRONE_TAG") and os.environ["DRONE_TAG"].removeprefix("v") != version:
        raise ValueError(f"Release tag must match the FyneApp.toml version/build: v{version}")
    environment = dict(os.environ, GOOS=args.target, GOARCH="amd64", CGO_ENABLED="1",
                       GOTOOLCHAIN="local", GOFLAGS="-mod=readonly")
    if args.target == "windows":
        environment.update(CC="x86_64-w64-mingw32-gcc", CXX="x86_64-w64-mingw32-g++")
    else:
        environment.update(CC="gcc", CXX="g++")
    flags = (f"-X whispering-tiger-ui/Updater.LinuxBackendFlavor={args.flavor} "
             f"-X whispering-tiger-ui/Updater.LinuxPreview={str(not release).lower()}")
    tags = ["migrated_fynedo"] if config.get("Migrations", {}).get("fyneDo") else []
    if release:
        tags.append("release")
    common = ["-buildvcs=false", "-ldflags", flags]
    if tags:
        common += ["-tags", ",".join(tags)]

    # Fyne's CLI uses this public API too. Generate only in an isolated build
    # copy; the Windows CLI remains free to generate its own metadata source.
    generated = source / "fyne_metadata_init.go"
    with generated.open("x", encoding="utf-8") as stream:
        stream.write('package main\nimport ("fyne.io/fyne/v2"; "fyne.io/fyne/v2/app"; "whispering-tiger-ui/Resources")\n')
        stream.write("func init() { app.SetMetadata(fyne.AppMetadata{\n")
        for key in ("ID", "Name", "Version", "Build"):
            stream.write(f"{key}: {json.dumps(details[key])},\n")
        stream.write(f"Release: {str(release).lower()}, Icon: Resources.ResourceAppIconPng,\n")
        custom = config.get("Release" if release else "Development", {})
        stream.write("Custom: map[string]string{" + ",".join(f"{json.dumps(k)}:{json.dumps(v)}" for k, v in custom.items()) + "},\n")
        stream.write("Migrations: map[string]bool{" + ",".join(f"{json.dumps(k)}:{str(v).lower()}" for k, v in config.get("Migrations", {}).items()) + "},\n}) }\n")
    try:
        run(["gofmt", "-w", str(generated)], source, environment)
        if args.test:
            if args.target == "linux":
                run(["xvfb-run", "-a", "go", "test", *common, "./..."], source, environment)
            else:
                print("Cross-compiling every Windows test package; tests are not executed on Linux.", flush=True)
                run(["go", "test", *common, "-exec", "/bin/true", "./..."], source, environment)
        args.output.parent.mkdir(parents=True, exist_ok=True)
        build_common = list(common)
        if args.target == "windows" and release:
            build_common[2] += " -H=windowsgui"
        run(["go", "build", *build_common, "-o", str(args.output), "."], source, environment)
        args.output.chmod(0o755)
        if args.target == "linux":
            result = subprocess.check_output(["xvfb-run", "-a", str(args.output), "--version"],
                                             cwd="/tmp", env=environment, text=True)
            if version not in result.splitlines():
                raise ValueError(f"Packaged executable did not report {version}: {result}")
            print(f"Verified executable metadata: {version}", flush=True)
    finally:
        generated.unlink()

    info = {"version": version, "os": args.target, "arch": "amd64", "release": release,
            "linux_backend_flavor": args.flavor, "go": subprocess.check_output(["go", "version"], text=True).strip()}
    info_path = args.output.parent / "build-info.json"
    info_path.write_text(json.dumps(info, indent=2) + "\n", encoding="utf-8")
    outputs = [args.output]
    if args.package:
        suffix = f"-linux-amd64-{args.flavor}" if args.target == "linux" else "-windows-amd64"
        archive = args.output.parent / f"whispering-tiger-ui-{version}{suffix}.zip"
        with zipfile.ZipFile(archive, "w", zipfile.ZIP_DEFLATED) as bundle:
            bundle.write(args.output, args.output.name)
            bundle.write(source / "LICENSE", "LICENSE")
            if args.target != "linux":
                bundle.write(info_path, info_path.name)
        outputs.append(archive)
    hashes = []
    for output in outputs:
        with output.open("rb") as stream:
            hashes.append(f"{hashlib.file_digest(stream, 'sha256').hexdigest()}  {output.name}\n")
    (args.output.parent / "SHA256SUMS").write_text("".join(hashes), encoding="utf-8")
    print(f"UI artifacts: {args.output.parent}", flush=True)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--source", type=Path, default=Path.cwd())
    parser.add_argument("--in-place", action="store_true", help="Only for an already isolated build copy")
    parser.add_argument("--target", choices=("linux", "windows"), default="linux")
    parser.add_argument("--flavor", choices=("cu128", "cpu"), default="cu128")
    parser.add_argument("--release", action="store_true", help="Enable Linux updates and release metadata")
    parser.add_argument("--release-if-tag", action="store_true", help="Drone tag events build releases; other events build previews")
    parser.add_argument("--test", action="store_true")
    parser.add_argument("--package", action="store_true")
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    args.source = args.source.resolve()
    args.output = args.output.resolve()
    if args.in_place:
        build(args.source, args)
    else:
        with tempfile.TemporaryDirectory(prefix="wt-ui-build-") as temporary:
            snapshot(args.source, Path(temporary))
            build(Path(temporary), args)


if __name__ == "__main__":
    main()
