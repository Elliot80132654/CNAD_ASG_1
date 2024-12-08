-- Create the database and switch to it
CREATE DATABASE IF NOT EXISTS electric_car_sharing;
USE electric_car_sharing;

-- Create the users table
CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    membership_id INT DEFAULT 1,  -- Add membership_id column directly in the table creation
    FOREIGN KEY (membership_id) REFERENCES memberships(id)  -- Link membership_id to the memberships table
);

-- Create the memberships table
CREATE TABLE IF NOT EXISTS memberships (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    hourly_rate_discount INT NOT NULL,
    vip_access BOOLEAN NOT NULL
);

-- Insert default memberships
INSERT INTO memberships (name, hourly_rate_discount, vip_access) VALUES
('Basic', 0, FALSE),
('Premium', 10, FALSE),
('VIP', 20, TRUE);

-- Create the user_details table
CREATE TABLE IF NOT EXISTS user_details (
    id INT PRIMARY KEY,  -- Foreign key linking to users table
    address VARCHAR(255) DEFAULT NULL,  -- User's address (nullable)
    phone_number VARCHAR(15) DEFAULT NULL,  -- User's phone number (nullable)
    gender ENUM('Male', 'Female', 'Other') DEFAULT NULL,  -- User's gender (nullable)
    FOREIGN KEY (id) REFERENCES users(id) ON DELETE CASCADE  -- Foreign key reference to users table
);

-- Create the vehicles table
CREATE TABLE IF NOT EXISTS vehicles (
    id INT AUTO_INCREMENT PRIMARY KEY,
    make VARCHAR(50) NOT NULL,
    model VARCHAR(50) NOT NULL,
    year INT NOT NULL,
    available BOOLEAN DEFAULT TRUE,
    vip_access BOOLEAN DEFAULT FALSE  -- Add vip_access directly in the vehicles table
);

-- Insert default vehicles
INSERT INTO vehicles (make, model, year, available, vip_access, cost_per_hour) VALUES
('Toyota', 'Corolla', 2020, TRUE, FALSE, 20.00),
('Honda', 'Civic', 2021, TRUE, FALSE, 20.00),
('Ford', 'Fiesta', 2019, TRUE, FALSE, 20.00),
('Chevrolet', 'Malibu', 2022, TRUE, FALSE, 20.00),
('Tesla', 'Model 3', 2023, TRUE, FALSE, 20.00),
('Hyundai', 'Elantra', 2020, TRUE, FALSE, 20.00),
('Tesla', 'Model S', 2023, TRUE, TRUE, 35.00),
('Porsche', 'Taycan', 2022, TRUE, TRUE, 35.00),
('BMW', 'i8', 2021, TRUE, TRUE, 35.00);

-- Create the rentals table
CREATE TABLE IF NOT EXISTS rentals (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT,
    vehicle_id INT,
    start_date DATETIME,
    end_date DATETIME,
    status ENUM('active', 'completed', 'cancelled') DEFAULT 'active',
    overtime_hours INT DEFAULT 0,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (vehicle_id) REFERENCES vehicles(id)
);

CREATE TABLE invoices (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    rental_id INT NOT NULL,
    hours INT NOT NULL,
    hours_overdue INT DEFAULT 0,
    final_cost DECIMAL(10, 2) NOT NULL,
    paid_status BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (rental_id) REFERENCES rentals(id),
    FOREIGN KEY (user_id) REFERENCES users(id)
);
