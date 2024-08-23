import api from './api.js';

// Populate address IDs dropdown
async function populateAddressIds() {
    const addressSelect = document.getElementById('addressIds');
    try {
        const addresses = await api.getAddresses();
        addresses.forEach(addr => {
            const option = document.createElement('option');
            option.value = addr.id;
            option.textContent = addr.display_title + ' ' + addr.display_subtitle;
            addressSelect.appendChild(option);
        });
    } catch (error) {
        console.error('Error fetching addresses:', error);
    }
}

// Time slot management
function addTimeSlot() {
    const timeSlotsContainer = document.getElementById('timeSlots');
    const timeSlotDiv = document.createElement('div');
    timeSlotDiv.classList.add('time-range');
    timeSlotDiv.innerHTML = `
        <input type="time" name="startTime[]" required>
        <span>to</span>
        <input type="time" name="endTime[]" required>
        <button type="button" class="removeTimeSlot">Remove</button>
    `;
    timeSlotsContainer.appendChild(timeSlotDiv);

    // Add event listener to the new "Remove" button
    const removeButton = timeSlotDiv.querySelector('.removeTimeSlot');
    removeButton.addEventListener('click', function () {
        this.parentElement.remove();
    });
}

const showAlert = (message) => {
    document.getElementById('alertText').textContent = message;
    document.getElementById('overlay').style.display = 'block';
    document.getElementById('customAlert').style.display = 'block';
};

document.getElementById('closeAlert').addEventListener('click', () => {
    document.getElementById('overlay').style.display = 'none';
    document.getElementById('customAlert').style.display = 'none';
});

// Form submission
async function handleSubmit(e) {
    e.preventDefault();
    const formData = new FormData(e.target);

    const apiPayload = {
        slackId: formData.get('slackId'),
        startTime: formData.getAll('startTime[]'),
        endTime: formData.getAll('endTime[]'),
        addressIds: formData.getAll('addressIds[]'),
    };

    const resp = await api.callWebhook(api.userSettingsEndpoint, apiPayload);
    console.log('Webhook response:', resp);

    if (resp !== "OK") {
        showAlert('Failed to save settings. Please try again later.')
        return;
    }

    showAlert('Settings saved successfully!\n You can check latest settings on slack using `/st-settings` command.');
    await chrome.storage.local.set({ snackTrackSlackId: formData.get('slackId') }, function () { });
    await await chrome.action.setBadgeText({ text: "ON" });
}

// Initialize the form
function init() {
    populateAddressIds();

    const addTimeSlotButton = document.getElementById('addTimeSlot');
    addTimeSlotButton.addEventListener('click', addTimeSlot);

    const form = document.getElementById('extensionForm');
    form.addEventListener('submit', handleSubmit);
}

// Run initialization when the DOM is fully loaded
document.addEventListener('DOMContentLoaded', init);
