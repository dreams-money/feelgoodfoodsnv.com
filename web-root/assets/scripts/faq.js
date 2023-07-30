const list = document.querySelector('.paginate');
if (list.children.length > 3) { // Ideally, calc screen length
   console.log(list.children);
   // Hide overflow
   for (let i = 3; i < list.children.length; i++) {
       list.children[i].style.display = 'none';
   }

   // Insert nav buttons
   const buttons = document.createElement('div')
   const previousButton = document.createElement('a');
   const nextButton = document.createElement('a');
   previousButton.setAttribute('href', '#');
   nextButton.setAttribute('href', '#');
   const previousGraphic = document.createElement('object');
   previousGraphic.setAttribute('type', 'image/svg+xml');
   previousGraphic.setAttribute('data', 'data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIGhlaWdodD0iNDgiIHZpZXdCb3g9IjAgOTYgOTYwIDk2MCIgd2lkdGg9IjQ4Ij48cGF0aCBkPSJNNDAwIDk3NiAwIDU3Nmw0MDAtNDAwIDU2IDU3LTM0MyAzNDMgMzQzIDM0My01NiA1N1oiLz48L3N2Zz4=')
   previousGraphic.setAttribute('style', 'pointer-events:none')
   const nextGraphic = document.createElement('object');
   nextGraphic.setAttribute('type', 'image/svg+xml');
   nextGraphic.setAttribute('data', 'data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIGhlaWdodD0iNDgiIHZpZXdCb3g9IjAgOTYgOTYwIDk2MCIgd2lkdGg9IjQ4Ij48cGF0aCBkPSJtMzA0IDk3NC01Ni01NyAzNDMtMzQzLTM0My0zNDMgNTYtNTcgNDAwIDQwMC00MDAgNDAwWiIvPjwvc3ZnPg==')
   nextGraphic.setAttribute('style', 'pointer-events:none')
   previousButton.appendChild(previousGraphic);
   nextButton.appendChild(nextGraphic);
   buttons.classList.add('page_btns')
   previousButton.classList.add('page_prev');
   nextButton.classList.add('page_next');
   let page = 0;
   flipPage = () => {
       for (let i = 0, pageMax = page + 3; i < list.children.length; i++) {
           if (i < page || i >= pageMax) {
               list.children[i].style.display = 'none';
           } else {
               list.children[i].style.display = 'block';
           }
       }
   };
   const flipPrevious = ()=>{
       if (page - 3 >= 0) {
           page = page - 3;
           flipPage();
       }
   }
   const flipNext = ()=>{
       if (page + 3 < list.children.length) {
           page = page + 3;
           flipPage();
       }
   };
   previousButton.addEventListener('click', flipPrevious);
   nextButton.addEventListener('click', flipNext);
   buttons.appendChild(previousButton);
   buttons.appendChild(nextButton);
   list.parentNode.insertBefore(buttons, list.nextSibling);
}