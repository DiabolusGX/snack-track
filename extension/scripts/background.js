import api from './api.js';

chrome.runtime.onInstalled.addListener(() => {
    console.log("Snack track extension installed");
    chrome.action.setBadgeText({
        text: "OFF",
    });

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

// poll running orders every minute
pollRunningOrders();
chrome.alarms.create("pollRunningOrders", { periodInMinutes: 1 });

chrome.alarms.onAlarm.addListener(async (alarm) => {
    console.log(`Alarm triggered for: ${alarm.name}`);

    const text = await new Promise((resolve, reject) => {
        chrome.action.getBadgeText({}, (text) => {
            if (chrome.runtime.lastError) {
                return reject(chrome.runtime.lastError);
            }
            resolve(text);
        });
    });
    if (text === "OFF") {
        return;
    }

    try {
        switch (alarm.name) {
            case "pollRunningOrders":
                await pollRunningOrders();
                break;
            default:
                break;
        }
    } catch (err) {
        console.error(err);
        showBasicNotification("login-cta", "Snack track ❌", `${err.message}. Please login and enable the extension again.`, [{ title: "Login" }]);
        await chrome.action.setBadgeText({
            text: "OFF",
        });
    }
});

async function pollRunningOrders() {
    const orders = await api.fetchOrders();
    if (orders.length == 0) {
        throw new Error("You have not placed any orders OR you are not logged in.");
    }

    let runningOrders = await new Promise((resolve, reject) => {
        chrome.storage.local.get("runningOrders", (data) => {
            if (chrome.runtime.lastError) {
                return reject(chrome.runtime.lastError);
            }
            resolve(data.runningOrders);
        });
    });
    if (!runningOrders) {
        runningOrders = [];
    }

    // check if `state` is changed for any running order
    const updatedOrders = orders.filter(order => {
        const runningOrder = runningOrders?.find(runningOrder => runningOrder.hashId === order.hashId);
        if (!runningOrder) {
            return false;
        }
        return runningOrder?.status !== order.status || runningOrder?.label !== order.deliveryDetails?.deliveryLabel;
    });
    for (const order of updatedOrders) {
        console.log(`Updated order: ${order.orderId} ${order.hashId}, ${order.status}, ${JSON.stringify(order.deliveryDetails)}`);
        const message = `[${order.orderId}] [${order.deliveryDetails?.deliveryLabel}] ${order.deliveryDetails?.deliveryLabel}`;
        await showBasicNotification("order-update", "Snack track 🚚", message);
        await api.notifyOnSlack(order);

        // if order is delivered, remove it from `runningOrders`
        if (!isRunningOrder(order)) {
            runningOrders = runningOrders?.filter(runningOrder => runningOrder.hashId !== order.hashId);
        }
    }

    // identify new orders that are not present in `runningOrders` and `state` is non-terminal
    const newOrders = orders.filter(order => {
        const runningOrder = runningOrders?.find(runningOrder => runningOrder.hashId === order.hashId);
        const isNewOrder = !runningOrder && isRunningOrder(order);
        if (isNewOrder) {
            console.log(`New order: ${order.orderId} ${order.hashId}, ${order.status}, ${JSON.stringify(order.deliveryDetails)}`);
            const message = `[${order.orderId}] [${order.deliveryDetails?.deliveryLabel}] ${order.deliveryDetails?.deliveryLabel}`;
            showBasicNotification("new-order", "Snack track 🚚", message);
            api.notifyOnSlack(order);
        }
        return isNewOrder;
    });

    // append new orders to `runningOrders`
    const finalRunningOrders = [...newOrders, ...updatedOrders];
    if (finalRunningOrders.length) {
        runningOrders = [];
        for (const order of finalRunningOrders) {
            runningOrders.push({
                hashId: order.hashId,
                status: order.status,
                label: order.deliveryDetails?.deliveryLabel,
            });
        }
    }

    chrome.storage.local.set({ runningOrders }, function () {
        console.log(`Running orders: ${JSON.stringify(runningOrders)}`);
    });
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
    return chrome.notifications.create(id, {
        type: "basic",
        iconUrl: "../public/logo512.png",
        title: title,
        message: message,
        buttons: buttons,
        requireInteraction: true,
    });
}
