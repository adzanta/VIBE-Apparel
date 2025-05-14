document.addEventListener("DOMContentLoaded", function() {
    const adaNamaKlub = document.getElementById("ada_nama_klub");
    const clubNameSection = document.getElementById("club_name_section");
    const jumlahPemainInput = document.getElementById("jumlah_pemain");
    const playerFormsDiv = document.getElementById("player_forms");
  
    adaNamaKlub.addEventListener("change", function() {
      if (this.value === "yes") {
        clubNameSection.style.display = "block";
      } else {
        clubNameSection.style.display = "none";
      }
    });
  
    jumlahPemainInput.addEventListener("input", function() {
      const jumlah = parseInt(this.value) || 0;
      playerFormsDiv.innerHTML = "";
  
      for (let i = 1; i <= jumlah; i++) {
        const playerDiv = document.createElement("div");
        playerDiv.classList.add("player-form");
        playerDiv.innerHTML = `
          <h4>Pemain ${i}</h4>
          <label>Nama Pemain: <input type="text" name="nama_pemain_${i}" required></label>
          <label>Nomor Punggung: <input type="text" name="nomor_punggung_${i}" required></label>
          <label>Ukuran: 
            <select name="ukuran_${i}" required>
              <option value="S">S</option>
              <option value="M">M</option>
              <option value="L">L</option>
              <option value="XL">XL</option>
              <option value="XXL">XXL</option>
            </select>
          </label>
        `;
        playerFormsDiv.appendChild(playerDiv);
      }
    });
  });
  