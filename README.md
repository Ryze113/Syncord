# Syncord deployment

This folder is the complete Syncord repository. It builds the Vencord source and publishes the files expected by the Syncord updater.

## Root server setup

Install Node.js 22+, pnpm, Go, and git. Clone or copy this workspace so these directories are next to each other:

```text
github/
  src/              # only on the build server, ignored by Git
    Vencord/
    Installer/
  build/
  settings/
```

Initialize this folder as the GitHub repository and configure authentication. Do not put a GitHub token in this script:

```sh
cd /root/github
git init
git remote add origin https://github.com/Ryze113/Syncord.git
git branch -M main
```

Make the script executable and run a build without publishing first:

```sh
chmod +x github/deploy.sh
PUSH=0 ./github/deploy.sh
```

The build output is written to `build`. The running updater downloads its files directly from the GitHub `build` directory. The complete source in `src` is intentionally ignored and is not published.

The source must remain on the server at `src/Vencord` and `src/Installer`, because `deploy.sh` builds from those directories. Do not delete `src` on the server.

To commit and push the source and published files to the configured `origin` remote:

```sh
./github/deploy.sh
```

The script currently builds the Linux amd64 CLI installer. To use different directories, set `SOURCE_DIR`, `SYNCORD_DIR`, or `OUTPUT_DIR` before running it.

## Remove source from GitHub

For an already initialized repository, remove `src` from Git tracking while keeping the files on the server:

```sh
git rm -r --cached src
git add .
git commit -m "Keep source private"
git push origin main
```
