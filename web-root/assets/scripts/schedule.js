const slots = document.querySelectorAll('.slot');
slots.forEach(slot => {
    slot.addEventListener('click', event => {
        const order = {
            slot_id: parseInt(event.target.getAttribute('slot-id')),
            items: JSON.parse(sessionStorage.getItem('order')),
            sub_total: parseFloat(sessionStorage.getItem('orderTotal')),
        };
        submitOrderForReview(order);
    })
});

function submitOrderForReview(order) {
    sessionStorage.setItem('slotID', order.slot_id);
    const orderSerialized = 'order='
        + encodeURIComponent(JSON.stringify(order));
    const url = '/review?' + orderSerialized;
    window.location = url;
}