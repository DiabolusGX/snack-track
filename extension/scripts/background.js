import api from './api.js';

chrome.runtime.onInstalled.addListener(() => {
    console.log("Snack track extension installed");
    chrome.action.setBadgeText({ text: "OFF" });

    chrome.storage.local.set({ runningOrders: [] }, function () {
        console.log(`Running orders reset`);
    });
});

chrome.action.onClicked.addListener(async (tab) => {
    console.log(`Extension clicked on tab: ${tab.id}`);

    chrome.tabs.sendMessage(tab.id, { type: "auth-status" }, async (response) => {
        await chrome.action.setBadgeText({
            text: response.isLoggedIn ? "ON" : "OFF",
        });
    });
});

// Poll running orders every minute
pollRunningOrders();
chrome.alarms.create("pollRunningOrders", { periodInMinutes: 1 });

chrome.alarms.onAlarm.addListener(async (alarm) => {
    console.log(`Alarm triggered for: ${alarm.name}`);

    try {
        const text = await getBadgeTextAsync();
        if (text === "OFF") {
            return;
        }

        switch (alarm.name) {
            case "pollRunningOrders":
                await pollRunningOrders();
                break;
            default:
                break;
        }
    } catch (err) {
        console.error(err);
        await showBasicNotification("login-cta", "Snack track ❌", `${err.message}. Please login and enable the extension again.`, [{ title: "Login" }]);
        await chrome.action.setBadgeText({ text: "OFF" });
    }
});

async function pollRunningOrders() {
    const slackId = await getStorageValueAsync("snackTrackSlackId");
    if (!slackId) throw new Error("Slack ID not set. Please set it in the 'Snack track' extension.");

    const orders = await api.fetchOrders();
    if (orders.length === 0) throw new Error("You have not placed any orders OR you are not logged in.");

    const runningOrders = await getStorageValueAsync("runningOrders") || [];

    const updatedOrders = [];
    const newOrders = [];
    const activeRunningOrders = [];

    for (const order of orders) {
        const runningOrder = runningOrders.find(ro => ro.hashId === order.hashId);
        if (runningOrder) {
            if (runningOrder.status !== order.status || runningOrder.label !== order.deliveryDetails?.deliveryLabel) {
                updatedOrders.push(order);
                continue
            }
            activeRunningOrders.push(order);
        } else if (isRunningOrder(order)) {
            newOrders.push(order);
        }
    }

    for (const order of updatedOrders) {
        console.log(`Updated order: ${order.orderId} ${order.hashId}, ${order.status}, ${JSON.stringify(order.deliveryDetails)}`);
        const message = `[${order.orderId}] [${order.deliveryDetails?.deliveryLabel}] ${order.deliveryDetails?.deliveryLabel}`;
        await showBasicNotification("order-update", "Snack track 🚚", message);
        await api.callWebhook(api.orderUpdateEndpoint, { order, slackId });
    }
    for (const order of newOrders) {
        console.log(`New order: ${order.orderId} ${order.hashId}, ${order.status}, ${JSON.stringify(order.deliveryDetails)}`);
        const message = `[${order.orderId}] [${order.deliveryDetails?.deliveryLabel}] ${order.deliveryDetails?.deliveryLabel}`;
        await showBasicNotification("new-order", "Snack track 🚚", message);
        await api.callWebhook(api.orderUpdateEndpoint, { order, slackId });
    }

    const updatedRunningOrders = updatedOrders.filter(isRunningOrder);
    const remainingRunningOrders = [...activeRunningOrders, ...newOrders, ...updatedRunningOrders];
    const finalRunningOrders = remainingRunningOrders.map(order => ({
        hashId: order.hashId,
        status: order.status,
        label: order.deliveryDetails?.deliveryLabel,
    }));

    await chrome.storage.local.set({ runningOrders: finalRunningOrders });
    console.log(`Running orders: ${JSON.stringify(finalRunningOrders)}`);
}

chrome.notifications.onButtonClicked.addListener((notificationId, buttonIndex) => {
    console.log(`Button clicked: ${notificationId}, buttonIndex: ${buttonIndex}`);
    if (notificationId === "login-cta") {
        chrome.tabs.create({ url: "https://www.zomato.com" });
    }
});

/******************** UTIL ********************/

function isRunningOrder(order) {
    return ![6, 7, 8].includes(order.status) && order.paymentStatus === 1;
}

function showBasicNotification(id, title, message, buttons = []) {
    return new Promise((resolve, reject) => {
        chrome.notifications.create(id, {
            type: "basic",
            iconUrl: "../public/logo512.png",
            title,
            message,
            buttons,
            requireInteraction: true,
        }, (notificationId) => {
            if (chrome.runtime.lastError) {
                return reject(chrome.runtime.lastError);
            }
            resolve(notificationId);
        });
    });
}

function getBadgeTextAsync() {
    return new Promise((resolve, reject) => {
        chrome.action.getBadgeText({}, (text) => {
            if (chrome.runtime.lastError) {
                return reject(chrome.runtime.lastError);
            }
            resolve(text);
        });
    });
}

function getStorageValueAsync(key) {
    return new Promise((resolve, reject) => {
        chrome.storage.local.get(key, (data) => {
            if (chrome.runtime.lastError) {
                return reject(chrome.runtime.lastError);
            }
            resolve(data[key]);
        });
    });
}
