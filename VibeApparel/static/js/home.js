document.addEventListener("DOMContentLoaded", function () {
    const burger = document.getElementById("burger-menu");
    const nav = document.getElementById("nav-links");
  
    burger.addEventListener("click", function () {
      nav.classList.toggle("show");
    });
  });
  