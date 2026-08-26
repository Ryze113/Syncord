import { showNotification } from "@api/Notifications";
import { Devs } from "@utils/constants";
import definePlugin from "@utils/types";

export default definePlugin({
    name: "Notify",
    description: "Shows a welcome notification",
    authors: [Devs.Ven],
    tags: ["Syncord"],
    start() {
        showNotification({
            title: "Welcome to Syncord :D!",
            body: "Syncord is running!",
            color: "#5865F2",
            noPersist: true
        });
    }
});
