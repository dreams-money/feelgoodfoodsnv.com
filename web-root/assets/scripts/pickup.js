const map = document.getElementById('map');
map.addEventListener('load', ()=>{
    document.getElementById('loading-spinner').style.display = 'none';
    map.style.display = 'block';
});

document.getElementById('address').addEventListener('click', () => {
    window.location = 'https://www.google.com/maps/dir//3922%20US-50,%20Carson%20City,%20NV%2089701';
});