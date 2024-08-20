chrome.runtime.onInstalled.addListener(() => {
    console.log("Snack track extension installed");
    chrome.action.setBadgeText({
        text: "OFF",
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
    const orders = await fetchOrders();
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
        return runningOrder?.status !== order.status;
    });
    for (const order of updatedOrders) {
        console.log(`Updated order: ${order.orderId} ${order.hashId}, ${order.status}, ${JSON.stringify(order.deliveryDetails)}`);
        const message = `[${order.orderId}] [${order.deliveryDetails?.deliveryLabel}] ${order.deliveryDetails?.deliveryLabel}`;
        await showBasicNotification("order-update", "Snack track 🚚", message);
        await notifyOnSlack(order);

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
        }
        return isNewOrder;
    });

    // append new orders to `runningOrders`
    const finalRunningOrders = [...newOrders, ...updatedOrders];
    runningOrders = [];
    for (const order of finalRunningOrders) {
        runningOrders.push({
            hashId: order.hashId,
            status: order.status,
        });
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

/******************** API ********************/

async function notifyOnSlack(order) {
    const webhookUrl = "https://snack-track.diabolus.me/webhook/order-update";
    const payload = JSON.stringify({ order });

    const requestOptions = {
        method: 'POST',
        mode: 'no-cors',
        headers: {
            'Content-Type': 'application/json',
        },
        body: payload,
    };

    return fetchWrrapper(webhookUrl, requestOptions);
}

async function fetchOrders() {
    const zh = await getZomatoHeaders();
    const requestOptions = {
        method: 'GET',
        headers: zh,
        redirect: 'follow'
    };

    const page = 0; // pagination not needed
    const response = await fetchWrrapper(`https://www.zomato.com/webroutes/user/orders?page=${page}`, requestOptions);
    if (response === "") {
        return {};
    }

    const ordersData = JSON.parse(response)?.entities?.ORDER;
    if (JSON.stringify(ordersData) === JSON.stringify([])) {
        return [];
    }

    const orderList = Object.keys(ordersData).map(orderId => ({
        id: orderId,
        ...ordersData[orderId]
    }));
    return orderList;
}

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

async function fetchWrrapper(url, options) {
    return fetch(url, options).then(response => {
        return response.text();
    }).catch(_ => {
        return "";
    });
}

async function getZomatoHeaders() {
    const cookies = await new Promise((resolve, reject) => {
        chrome.cookies.getAll({ url: "https://www.zomato.com" }, (cookies) => {
            if (chrome.runtime.lastError) {
                return reject(chrome.runtime.lastError);
            }
            resolve(cookies);
        });
    });

    const cookieMap = {};
    cookies.forEach(cookie => {
        cookieMap[cookie.name] = cookie.value;
    });
    const cookieString = Object.keys(cookieMap).map(key => `${key}=${cookieMap[key]}`).join('; ');

    const zh = new Headers();
    zh.append("authority", "www.zomato.com");
    zh.append("sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"99\", \"Google Chrome\";v=\"99\"");
    zh.append("x-zomato-csrft", cookieMap?.csrf);
    zh.append("sec-ch-ua-mobile", "?1");
    zh.append("user-agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/99.0.4844.82 Mobile Safari/537.36");
    zh.append("sec-ch-ua-platform", "\"Android\"");
    zh.append("accept", "*/*");
    zh.append("sec-fetch-site", "same-origin");
    zh.append("sec-fetch-mode", "cors");
    zh.append("sec-fetch-dest", "empty");
    zh.append("accept-language", "en-US,en;q=0.9");
    zh.append("cookie", cookieString);
    return zh;
}
