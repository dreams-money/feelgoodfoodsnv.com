const maxAllowableQuantity = 5;

const menuItems = document.querySelectorAll('.menu_item');
menuItems.forEach(item => {
    item.addEventListener('click', () => {
        item.style.borderColor = 'var(--border-visible)';
        item.querySelector('.quantity_control').style.opacity = 1;
        if (parseInt(item.querySelector('.quantity').innerHTML) == 0) {
            item.querySelector('.quantity').innerHTML = 1
        }
    });
});

const addButtons = document.querySelectorAll('.add');
addButtons.forEach(button => {
    button.addEventListener('click', event => {
        event.stopPropagation();
        const menuItem = parentMenuItemRelativeToKeyPress(event)
        menuItem.style.borderColor = 'var(--border-visible)';
        const quantity = button.parentElement.querySelector('.quantity');
        const currentAmount = parseInt(quantity.innerHTML);
        if (currentAmount < maxAllowableQuantity) {
            quantity.innerHTML = currentAmount + 1;
        }
    });
});

const subtractButtons = document.querySelectorAll('.subtract');
subtractButtons.forEach(button => {
    button.addEventListener('click', (event) => {
        event.stopPropagation();
        const quantity = button.parentElement.querySelector('.quantity');
        const currentAmount = parseInt(quantity.innerHTML);
        if (currentAmount > 1) {
            quantity.innerHTML = currentAmount - 1;
        } else {
            unselectMenuItemWithClickEvent(event);
        }
    });
});

const protienButtons = document.querySelectorAll('.protien_btn');
protienButtons.forEach(button => {
    button.addEventListener('click', event => {
        const img = event.target.querySelector('img');
        if (img == null) {
            return;
        }
        if (event.target.selected) {
            event.target.selected = false;
            img.src = '/assets/icons/x-solid.svg';
        } else {
            event.target.selected = true;
            img.src = '/assets/icons/check-solid.svg';
        }
    });
});

const submitButton = document.getElementById('submit_btn');
if (submitButton != null && submitButton.tagName.toLowerCase() != 'div') {
    submitButton.addEventListener('click', () => {
        const orders = [];
        let orderTotal = 0;
        menuItems.forEach(menuItem => {
            let extraProtien = false;
            if (menuItem.querySelector('.protien_btn').selected) {
                extraProtien = menuItem.querySelector('.protien_btn').selected;
            }
            const orderItem = {
                menu_item: {
                    id: parseInt(menuItem.getAttribute('menu-id')),
                },
                quantity: parseInt(menuItem.querySelector('.quantity').innerHTML),
                price: parseFloat(menuItem.querySelector('.item_price').innerHTML),
                extra_protien: extraProtien,
            };
            if (orderItem.quantity > 0) {
                orders.push(orderItem);
                orderTotal += orderItem.quantity * orderItem.price;
            }
        });

        if (orderTotal == 0) {
            alert("Please add items to your order.");
        } else if (orderTotal < 50) {
            alert("Sorry, your order must be at least $50 to be submitted.");
        } else {
            saveOrder(orders, orderTotal);
        }
    })
}

const parentMenuItemRelativeToKeyPress = (event) => {
    return event.target.parentElement.parentElement.parentElement;
}

const unselectMenuItemWithClickEvent = event => {
    const menuItem = parentMenuItemRelativeToKeyPress(event)
    const quantity = menuItem.querySelector('.quantity')
    const quantityControl = menuItem.querySelector('.quantity_control');

    quantity.innerHTML = 0;
    // quantityControl.style.opacity = 0;
    menuItem.style.borderColor = 'var(--border-nosee)';
}

const saveOrder = (orders, orderTotal) => {
    sessionStorage.setItem('order', JSON.stringify(orders))
    sessionStorage.setItem('orderTotal', orderTotal)
    window.location = '/schedule';
}