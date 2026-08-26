/*
 * Syncord, a modification for Discord's desktop app
 * Copyright (c) 2023 Vendicated and contributors
 */

import { fetchBuffer } from "@main/utils/http";
import { DATA_DIR } from "@main/utils/constants";
import { IpcEvents } from "@shared/IpcEvents";
import { BrowserWindow, ipcMain } from "electron";
import { readFile, writeFile, unlink } from "fs/promises";
import { join } from "path";
import { existsSync } from "fs";

import { serializeErrors, VENCORD_FILES } from "./common";

const VERSION_URL = "https://raw.githubusercontent.com/Ryze113/Syncord/main/build/version.txt";
const DIST_URL = "https://raw.githubusercontent.com/Ryze113/Syncord/main/build/";
const LOCAL_VERSION_FILE = join(DATA_DIR, "version.txt");

let remoteVersion = "";
let localVersion = typeof SYNCORD_VERSION !== "undefined" ? SYNCORD_VERSION : "1.0.0";
let isOutdated = false;
let lastCheck = "";
let currentStatus: "checking" | "up-to-date" | "update-available" | "error" = "checking";

function notifyRenderer(channel: string, ...args: any[]) {
    BrowserWindow.getAllWindows().forEach(w => {
        if (!w.isDestroyed()) {
            w.webContents.send(channel, ...args);
        }
    });
}

async function loadLocalVersion() {
    try {
        if (existsSync(LOCAL_VERSION_FILE)) {
            const saved = (await readFile(LOCAL_VERSION_FILE, "utf-8")).trim();
            if (saved) localVersion = saved;
        }
    } catch {}
}

async function saveLocalVersion(version: string) {
    try { await writeFile(LOCAL_VERSION_FILE, version, "utf-8"); } catch {}
}

async function checkForUpdates() {
    currentStatus = "checking";
    notifyRenderer("VencordSyncordStatus", currentStatus, localVersion, remoteVersion, lastCheck);
    try {
        const res = await fetch(VERSION_URL, { cache: "no-store" } as any);
        if (!res.ok) throw new Error(`HTTP ${res.status}`);
        remoteVersion = (await res.text()).trim();
        lastCheck = new Date().toLocaleTimeString();
        isOutdated = remoteVersion !== "" && localVersion !== "" && remoteVersion !== localVersion;
        currentStatus = isOutdated ? "update-available" : "up-to-date";
        if (isOutdated) notifyRenderer("VencordSyncordUpdate", remoteVersion, localVersion);
    } catch { currentStatus = "error"; }
    notifyRenderer("VencordSyncordStatus", currentStatus, localVersion, remoteVersion, lastCheck);
}

async function downloadFiles(): Promise<boolean> {
    try {
        for (const fileName of VENCORD_FILES) {
            const buffer = await fetchBuffer(DIST_URL + fileName);
            const dest = join(__dirname, fileName);
            try { if (existsSync(dest)) await unlink(dest); } catch {}
            await writeFile(dest, buffer);
        }
        try {
            const vRes = await fetch(VERSION_URL, { cache: "no-store" } as any);
            if (vRes.ok) {
                const vText = (await vRes.text()).trim();
                try { await writeFile(join(__dirname, "version.txt"), vText, "utf-8"); } catch {}
            }
        } catch {}
        return true;
    } catch (err) {
        console.error("[Syncord] Download failed:", err);
        return false;
    }
}

async function applyUpdate(): Promise<boolean> {
    if (!isOutdated) return false;
    const success = await downloadFiles();
    if (success) {
        await saveLocalVersion(remoteVersion);
        localVersion = remoteVersion;
        isOutdated = false;
    }
    return success;
}

ipcMain.handle(IpcEvents.GET_UPDATES, serializeErrors(async () => {
    if (isOutdated) return [{ hash: remoteVersion, author: "Syncord", message: `Update to ${remoteVersion}` }];
    return [];
}));
ipcMain.handle(IpcEvents.UPDATE, serializeErrors(applyUpdate));
ipcMain.handle(IpcEvents.BUILD, serializeErrors(async () => true));
ipcMain.handle(IpcEvents.GET_REPO, serializeErrors(() => "https://github.com/Ryze113/Syncord"));
ipcMain.handle(IpcEvents.GET_LOCAL_VERSION, serializeErrors(async () => localVersion));

(async () => {
    await loadLocalVersion();
    await checkForUpdates();
    setInterval(checkForUpdates, 10_000);
})();
