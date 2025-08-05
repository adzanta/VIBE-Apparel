# Custom Jersey Order (VIBE Apparel)

A pre-order jersey customization website developed for **VIBE Apparel**. This project uses **Go (Golang)** as the backend and pure **HTML/CSS/JavaScript** for the frontend. It supports custom design submission, order tracking, and an admin dashboard for managing incoming orders.

## ✨ Features

- Custom jersey design form
- Admin dashboard for managing orders
- Order status tracking by user
- Secure login with hashed passwords
- Pre-order system with estimated delivery

## 🛠 Tech Stack

- **Backend**: Go (Echo framework)
- **Frontend**: HTML, CSS, JavaScript
- **Database**: MySQL (XAMPP)
- **Templating**: HTML templates rendered by Go
- **Authentication**: Bcrypt password hashing, sessions, and role-based access (admin/user)

## 📁 Project Structure (Simplified)

```
/vibe-apparel
├── main.go
├── /routes
├── /controllers
├── /models
├── /templates
├── /static
│   ├── /css
│   └── /js
└── /database
```

## 🚀 How to Run Locally

1. Clone the repository:

```bash
git clone https://github.com/adzanta/vibe-apparel.git
cd vibe-apparel
```

2. Start MySQL server (e.g. via XAMPP) and import the database schema from `/database/schema.sql`.

3. Configure your `.env` file or database credentials directly in the source.

4. Run the Go server:

```bash
go run main.go
```

5. Visit `http://localhost:8080` in your browser.

## 📌 Notes

- Pre-order only, no live payment integration
- Custom design feature allows users to submit special requests
- Admin dashboard includes order management and status updates

## 👤 Author

Alhafidz William Adzanta  
[LinkedIn](https://www.linkedin.com/in/alhafidz-william-adzanta/)

---

