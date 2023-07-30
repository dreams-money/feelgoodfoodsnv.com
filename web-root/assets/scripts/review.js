'use strict';

document.addEventListener('DOMContentLoaded', ()=>{
    document.getElementById('pay_btn').addEventListener('click', () => {
        window.location = '/pay' + window.location.search;
    });
});