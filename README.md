# Syncord deployment

This folder is the complete Syncord repository. It builds the Vencord source and publishes the files expected by the Syncord updater.

## Root server setup

Install Node.js 22+, pnpm, Go, and git. Clone or copy this workspace so these directories are next to each other:

```text
github/
  src/Vencord/
  src/Installer/
  build/
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

The build output is written to `build`. The running updater downloads its files directly from the GitHub `build` directory.

To commit and push the source and published files to the configured `origin` remote:

```sh
./github/deploy.sh
```

The script currently builds the Linux amd64 CLI installer. To use different directories, set `SOURCE_DIR`, `SYNCORD_DIR`, or `OUTPUT_DIR` before running it.
