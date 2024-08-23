import api from './api.js';

// Populate address IDs dropdown
async function populateAddressIds() {
    const addressSelect = document.getElementById('addressIds');
    try {
        const addresses = await api.getAddresses();
        addresses.forEach(addr => {
            const option = document.createElement('option');
            option.value = addr.id;
            option.textContent = addr.alias;
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
    timeSlotDiv.innerHTML = `
        <input type="time" name="timeSlot[]" required>
        <button type="button" class="removeTimeSlot">Remove</button>
    `;
    timeSlotsContainer.appendChild(timeSlotDiv);

    // Add event listener to the new "Remove" button
    const removeButton = timeSlotDiv.querySelector('.removeTimeSlot');
    removeButton.addEventListener('click', function () {
        this.parentElement.remove();
    });
}

// Form submission
function handleSubmit(e) {
    e.preventDefault();
    const formData = new FormData(e.target);
    const addressSelect = document.getElementById('addressIds');
    const selectedAddressIds = Array.from(addressSelect.selectedOptions).map(option => option.value);
    const timeSlots = formData.getAll('timeSlot[]');

    console.log('Selected Address IDs:', selectedAddressIds);
    console.log('Time Slots:', timeSlots);

    // Here you would typically send this data to your backend or process it further
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
