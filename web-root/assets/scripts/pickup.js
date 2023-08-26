const map = document.getElementById('map');
map.addEventListener('load', ()=>{
    document.getElementById('loading-spinner').style.display = 'none';
    map.style.display = 'block';
});