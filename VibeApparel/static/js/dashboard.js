document.addEventListener('DOMContentLoaded', function() {
    // Toggle menu untuk tampilan mobile
    const menuButton = document.createElement('button');
    menuButton.classList.add('menu-toggle');
    menuButton.innerHTML = '<span></span><span></span><span></span>';
    document.querySelector('.main-header').prepend(menuButton);
    
    // Tambahkan CSS inline untuk menu toggle
    const style = document.createElement('style');
    style.textContent = `
        .menu-toggle {
            display: none;
            background: none;
            border: none;
            cursor: pointer;
            margin-right: 15px;
            padding: 10px;
        }
        
        .menu-toggle span {
            display: block;
            width: 25px;
            height: 3px;
            background-color: var(--text-color);
            margin: 5px 0;
            transition: var(--transition);
        }
        
        @media (max-width: 767px) {
            .menu-toggle {
                display: block;
            }
            
            .main-header {
                display: flex;
                align-items: center;
            }
            
            .menu-toggle.active span:nth-child(1) {
                transform: rotate(45deg) translate(5px, 5px);
            }
            
            .menu-toggle.active span:nth-child(2) {
                opacity: 0;
            }
            
            .menu-toggle.active span:nth-child(3) {
                transform: rotate(-45deg) translate(7px, -7px);
            }
        }
    `;
    document.head.appendChild(style);
    
    // Event listener untuk toggle menu pada tampilan mobile
    menuButton.addEventListener('click', function() {
        this.classList.toggle('active');
        document.querySelector('.sidebar').classList.toggle('active');
    });
    
    // Tambahkan font Poppins dari Google Fonts
    const poppinsFont = document.createElement('link');
    poppinsFont.rel = 'stylesheet';
    poppinsFont.href = 'https://fonts.googleapis.com/css2?family=Poppins:wght@300;400;500;600;700&display=swap';
    document.head.appendChild(poppinsFont);
    
    // Tambahkan Font Awesome untuk ikon
    const fontAwesome = document.createElement('link');
    fontAwesome.rel = 'stylesheet';
    fontAwesome.href = 'https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css';
    document.head.appendChild(fontAwesome);
    
    // Fix double text in menu items
    const menuItems = document.querySelectorAll('.sidebar-menu a');
    menuItems.forEach(item => {
        const text = item.textContent.trim();
        // Cek jika teks duplikat (seperti "Dashboard Dashboard")
        if (text.includes(" ")) {
            const firstWord = text.split(" ")[0]; // Ambil kata pertama saja
            
            // Tambahkan ikon sesuai dengan menu
            let iconClass = '';
            if (firstWord.includes('Dashboard')) {
                iconClass = 'fas fa-th-large';
            } else if (firstWord.includes('Produk')) {
                iconClass = 'fas fa-tshirt';
            } else if (firstWord.includes('Pesanan')) {
                iconClass = 'fas fa-shopping-cart';
            }
            
            // Update konten dengan ikon dan teks yang benar
            item.innerHTML = `<i class="${iconClass}"></i> <span class="menu-text">${firstWord}</span>`;
        }
    });
    
    // Animasi counter untuk nilai pada kartu statistik
    const statValues = document.querySelectorAll('.card p');
    
    statValues.forEach(value => {
        const finalValue = value.innerText;
        let startValue = 0;
        
        // Jika nilai berisi Rp, animasikan sebagai nilai mata uang
        if (finalValue.includes('Rp')) {
            const numericValue = parseInt(finalValue.replace(/[^\d]/g, ''));
            const duration = 1500;
            const increment = numericValue / (duration / 20);
            
            const counter = setInterval(() => {
                startValue += increment;
                if (startValue >= numericValue) {
                    value.innerText = finalValue;
                    clearInterval(counter);
                } else {
                    value.innerText = 'Rp ' + Math.floor(startValue).toLocaleString('id-ID');
                }
            }, 20);
        } else {
            // Jika nilai hanya berupa angka
            const numericValue = parseInt(finalValue);
            const duration = 1000;
            const increment = numericValue / (duration / 20);
            
            const counter = setInterval(() => {
                startValue += increment;
                if (startValue >= numericValue) {
                    value.innerText = finalValue;
                    clearInterval(counter);
                } else {
                    value.innerText = Math.floor(startValue);
                }
            }, 20);
        }
    });
});