document.addEventListener("DOMContentLoaded", function() {
    const button = document.querySelector(".checkout-form button");
    button.addEventListener("mouseenter", function() {
      button.style.backgroundColor = "#388E3C";
    });
    button.addEventListener("mouseleave", function() {
      button.style.backgroundColor = "#4CAF50";
    });
  });
  