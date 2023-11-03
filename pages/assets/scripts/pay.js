{{define "pay.js"}}
{{if eq .Environment "dev"}}
const appId = '{{.DevPublicPayKey}}';
{{else}}
const appId = '{{.ProdPublicPayKey}}';
{{end}}
const locationId = 'LMMWS17R0TK2J';

async function initializePaymentFormAndCard(payments) {
  const card = await payments.card({
    style: {
      'input': {
        backgroundColor: '#E5D0B8',
        color: 'black',
        fontFamily: 'Georgia, Times New Roman, Times, serif',
        fontWeight: '400',
      },
      '.input-container.is-focus': {
        borderColor: 'black'
      },
      'input::placeholder': {
        color: 'black'
      },
      '.message-text': {
        color: 'black'
      },
      '.message-icon': {
        color: 'black'
      }
    }
  });
  await card.attach('#card-container');

  return card;
}

async function fetchTokenFromPaymentProvider(paymentMethod) {
  const tokenResult = await paymentMethod.tokenize();
  if (tokenResult.status === 'OK') {
    return tokenResult.token;
  } else {
    let errorMessage = `Tokenization failed with status: ${tokenResult.status}`;
    if (tokenResult.errors) {
      errorMessage += ` and errors: ${JSON.stringify(
        tokenResult.errors
      )}`;
    }

    throw new Error(errorMessage);
  }
}

async function submitOrderToAppServer(token, order) {
  const body = JSON.stringify({
    location: locationId,
    token,
    order
  });

  const paymentResponse = await fetch('/submit', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body,
  });

  if (paymentResponse.ok) {
    return paymentResponse.json();
  }

  const errorBody = await paymentResponse.text();
  throw new Error(errorBody);
}

// status is either SUCCESS or FAILURE;
function displayPaymentResults(status) {
  const statusContainer = document.getElementById(
    'payment-status-container'
  );
  if (status === 'SUCCESS') {
    statusContainer.classList.remove('is-failure');
    statusContainer.classList.add('is-success');
  } else {
    statusContainer.classList.remove('is-success');
    statusContainer.classList.add('is-failure');
  }

  statusContainer.style.visibility = 'visible';
}

async function getAllowedZipCodes(slotId) {
  const paymentResponse = await fetch('/allowed-zips?slot-id='+slotId);

  if (paymentResponse.ok) {
    return paymentResponse.json();
  }

  const errorBody = await paymentResponse.text();
  throw new Error(errorBody);
}

function getCustomer() {
  const form = document.getElementById('customer_form');
  const inputs = form.querySelectorAll('input')
  const customer = {}
  for (let i = 0; i < inputs.length; i++) {
    customer[inputs[i].name] = inputs[i].value;
  }

  if (customer.hasOwnProperty('postal')) {
    customer["addresses"] = [
      {
        type: "delivery",
        street1: customer.street1,
        street2: customer.street2,
        postal: customer.postal,
      }
    ];

    delete customer.street1;
    delete customer.street2;
    delete customer.postal;
  }

  return customer;
}

function getOrder(customer) {
  return {
    items: JSON.parse(sessionStorage.getItem('order')),
    sub_total: parseFloat(sessionStorage.getItem('orderTotal')),
    customer,
    slot_id: parseInt(sessionStorage.getItem('slotID')),
  };
}

function getOrderFromURL() {
  return JSON.parse(decodeURIComponent(window.location.search.replace('?order=','')))
}

function putOrderBackInURL(order) {
  return 'order=' + encodeURIComponent(JSON.stringify(order));
}

document.addEventListener('DOMContentLoaded', async () => {
  if (!window.Square) {
    throw new Error('Square.js failed to load properly');
  }
  const statusContainer = document.getElementById(
    'payment-status-container'
  );

  let payments;
  try {
    payments = window.Square.payments(appId, locationId);
  } catch (e) {
    console.error('Invalid API Credentials', e);
    return;
  }

  let card;
  try {
    card = await initializePaymentFormAndCard(payments);
  } catch (e) {
    console.error('Initializing Card failed', e);
    return;
  }

  // Checkpoint 2
  async function handlePaymentMethodSubmission(event, paymentMethod) {
    event.preventDefault();

    try {
      // disable the submit button as we await tokenization and make a payment request.
      cardButton.disabled = true;

      const order = getOrder(getCustomer())
      const token = await fetchTokenFromPaymentProvider(paymentMethod);

      const orderResponse = await submitOrderToAppServer(token, order);
      if (orderResponse.payment_status !== "COMPLETED") {
        throw new Error('Payment failed!');
      }

      const orderCached = getOrderFromURL();
      orderCached.id = orderResponse.new_order_id;
      const newQueryURL = putOrderBackInURL(orderCached)

      displayPaymentResults('SUCCESS');
      setTimeout(() => {
        window.location = '/receipt?' + newQueryURL
      }, 5000)
    } catch (e) {
      cardButton.disabled = false;
      displayPaymentResults('FAILURE');
      console.error(e.message);
    }
  }

  const cardButton = document.getElementById('card-button');
  cardButton.addEventListener('click', async function (event) {
    const form = document.getElementById('customer_form');
    if (form.reportValidity()) {
      await handlePaymentMethodSubmission(event, card);
    }
  });

  const cellNumber = document.getElementById('phone');
  cellNumber.addEventListener('click', event => {
    event.stopPropagation();
    event.target.setAttribute('placeholder', '0000000000');
    document.querySelector('html').addEventListener('click', e => {
      cellNumber.setAttribute('placeholder', 'Cell phone number');
      e.target.removeEventListener('click');
    });
  });

  const allowedZips = await getAllowedZipCodes(parseInt(sessionStorage.getItem('slotID')));
  if (allowedZips != null) {
    document.getElementById('postal').addEventListener('blur', async e => {
      const inputValue = e.target.value;
      let inputZip = 0;
      if (inputValue.includes('-')) {
        [inputZip] = inputValue.split('-');
      } else {
        inputZip = inputValue
      }
      inputZip = parseInt(inputZip);

      if (inputZip > 0 && !allowedZips.includes(inputZip)) {
        const m = "Sorry, the selected delivery slot doesn't match your delivery address."
        alert(m);
      }
    });
  }
})
{{end}}