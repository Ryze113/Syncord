/*
 * Vencord, a modification for Discord's desktop app
 * Copyright (c) 2022 Vendicated and contributors
 */

import { Card } from "@components/Card";
import { Divider } from "@components/Divider";
import { Flex } from "@components/Flex";
import { Link } from "@components/Link";
import { SettingsTab, wrapTab } from "@components/settings/tabs/BaseTab";
import { Forms } from "@webpack/common";
import { React } from "@webpack/common";

function SyncordUpdater() {
    const [status, setStatus] = React.useState<"checking" | "up-to-date" | "update-available" | "error">("checking");
    const [localVersion, setLocalVersion] = React.useState("...");
    const [remoteVersion, setRemoteVersion] = React.useState("...");
    const [lastCheck, setLastCheck] = React.useState("...");
    const [nextCheck, setNextCheck] = React.useState(10);
    React.useEffect(() => {
        VencordNative.updater.getLocalVersion().then(res => {
            if (res.ok) setLocalVersion(res.value);
        }).catch(() => {});
    }, []);
    React.useEffect(() => {
        let countdownInterval: ReturnType<typeof setInterval>;
        VencordNative.updater.onUpdateAvailable((remote: string, local: string) => {
            setRemoteVersion(remote);
            setLocalVersion(local);
            setStatus("update-available");
            setLastCheck(new Date().toLocaleTimeString());
            setNextCheck(10);
        });
        VencordNative.updater.onStatusUpdate((newStatus: string, local: string, remote: string, last: string) => {
            setStatus(newStatus as any);
            if (local) setLocalVersion(local);
            if (remote) setRemoteVersion(remote);
            if (last) setLastCheck(last);
            setNextCheck(10);
        });
        countdownInterval = setInterval(() => {
            setNextCheck(prev => (prev > 0 ? prev - 1 : 0));
        }, 1000);
        return () => { clearInterval(countdownInterval); };
    }, []);
    const statusColors = {
        "checking": "var(--info-warning-foreground)",
        "up-to-date": "var(--green-360)",
        "update-available": "var(--yellow-360)",
        "error": "var(--red-360)",
    };
    const statusText = {
        "checking": "Checking...",
        "up-to-date": "Up to date",
        "update-available": "Update available!",
        "error": "Failed to check",
    };
    return (
        <SettingsTab>
            <Card variant="info" style={{ padding: "1em" }}>
                <Flex flexDirection="column" gap="0.8em">
                    <Flex alignItems="center" gap="0.5em">
                        <div style={{ width: 10, height: 10, borderRadius: "50%", backgroundColor: statusColors[status], display: "inline-block" }} />
                        <Forms.FormText style={{ fontWeight: 600 }}>Syncord Updater: {statusText[status]}</Forms.FormText>
                    </Flex>
                    <Divider />
                    <Forms.FormText>Built Version: <b>{localVersion}</b></Forms.FormText>
                    <Forms.FormText>Remote Version: <b>{remoteVersion}</b></Forms.FormText>
                    <Forms.FormText>Last Check: <b>{lastCheck}</b></Forms.FormText>
                    <Forms.FormText>Next Check: <b>{nextCheck}s</b></Forms.FormText>
                    <Divider />
                    <Forms.FormText>Polls every 10s from <Link href="https://github.com/Ryze113/Syncord">Ryze113/Syncord</Link></Forms.FormText>
                </Flex>
            </Card>
            {status === "update-available" && (
                <Card variant="warning" style={{ padding: "1em", marginTop: "1em" }}>
                    <Forms.FormText style={{ fontWeight: 600 }}>A new Syncord version is available! Click the notification at the bottom right to install.</Forms.FormText>
                </Card>
            )}
        </SettingsTab>
    );
}
export default wrapTab(SyncordUpdater, "Updater");
