console.log("Snack track extension API script loaded");

export default {
    getAddresses,
};

async function getAddresses() {
    // TODO: make actual API request
    return [
        {
            "id": 123,
            "address": "Zomato, 2nd Floor, Tower 1",
            "alias": "Work"
        },
        {
            "id": 456,
            "address": "My home",
            "alias": "Home"
        }
    ];
}
