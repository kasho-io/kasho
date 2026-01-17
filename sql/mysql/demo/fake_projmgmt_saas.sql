-- Fake Project Management SaaS data for MySQL
-- Compatible with MySQL 8.0+

SET FOREIGN_KEY_CHECKS = 0;

-- Organizations table
DROP TABLE IF EXISTS organizations;
CREATE TABLE organizations (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,
    created_at DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6)
) ENGINE=InnoDB;

-- Users table
DROP TABLE IF EXISTS users;
CREATE TABLE users (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    organization_id CHAR(36) NOT NULL,
    email VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    password_hash VARCHAR(255),
    created_at DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    UNIQUE KEY idx_email (email)
) ENGINE=InnoDB;

-- Credit cards table (sensitive data)
DROP TABLE IF EXISTS credit_cards;
CREATE TABLE credit_cards (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    organization_id CHAR(36) NOT NULL,
    card_number VARCHAR(20) NOT NULL,
    expiry_month TINYINT NOT NULL,
    expiry_year SMALLINT NOT NULL,
    cvv VARCHAR(4) NOT NULL,
    cardholder_name VARCHAR(255) NOT NULL,
    created_at DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6),
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
) ENGINE=InnoDB;

-- Subscriptions table
DROP TABLE IF EXISTS subscriptions;
CREATE TABLE subscriptions (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    organization_id CHAR(36) NOT NULL,
    plan_name VARCHAR(50) NOT NULL,
    status ENUM('active', 'canceled', 'past_due', 'trialing') DEFAULT 'active',
    started_at DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6),
    ends_at DATETIME(6),
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
) ENGINE=InnoDB;

-- Invoices table
DROP TABLE IF EXISTS invoices;
CREATE TABLE invoices (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    organization_id CHAR(36) NOT NULL,
    amount_cents INT NOT NULL,
    currency CHAR(3) DEFAULT 'USD',
    status ENUM('draft', 'sent', 'paid', 'void') DEFAULT 'draft',
    due_date DATE,
    paid_at DATETIME(6),
    created_at DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6),
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
) ENGINE=InnoDB;

-- Projects table
DROP TABLE IF EXISTS projects;
CREATE TABLE projects (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    organization_id CHAR(36) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status ENUM('active', 'archived', 'completed') DEFAULT 'active',
    created_at DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
) ENGINE=InnoDB;

-- Tasks table
DROP TABLE IF EXISTS tasks;
CREATE TABLE tasks (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    project_id CHAR(36) NOT NULL,
    assignee_id CHAR(36),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status ENUM('todo', 'in_progress', 'review', 'done') DEFAULT 'todo',
    priority TINYINT DEFAULT 2,
    due_date DATE,
    created_at DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY (assignee_id) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB;

SET FOREIGN_KEY_CHECKS = 1;

-- Insert sample data
-- Organizations
INSERT INTO organizations (id, name, slug) VALUES
    ('11111111-1111-1111-1111-111111111111', 'Acme Corporation', 'acme'),
    ('22222222-2222-2222-2222-222222222222', 'TechStart Inc', 'techstart'),
    ('33333333-3333-3333-3333-333333333333', 'Global Solutions', 'global-solutions');

-- Users (with sensitive data that would be transformed)
INSERT INTO users (id, organization_id, email, first_name, last_name, password_hash) VALUES
    ('aaaa1111-1111-1111-1111-111111111111', '11111111-1111-1111-1111-111111111111', 'john.doe@acme.com', 'John', 'Doe', '$2b$12$hashedpassword1'),
    ('aaaa2222-2222-2222-2222-222222222222', '11111111-1111-1111-1111-111111111111', 'jane.smith@acme.com', 'Jane', 'Smith', '$2b$12$hashedpassword2'),
    ('bbbb1111-1111-1111-1111-111111111111', '22222222-2222-2222-2222-222222222222', 'bob@techstart.io', 'Bob', 'Johnson', '$2b$12$hashedpassword3');

-- Credit cards (highly sensitive)
INSERT INTO credit_cards (id, organization_id, card_number, expiry_month, expiry_year, cvv, cardholder_name) VALUES
    ('cc111111-1111-1111-1111-111111111111', '11111111-1111-1111-1111-111111111111', '4111111111111111', 12, 2026, '123', 'John Doe'),
    ('cc222222-2222-2222-2222-222222222222', '22222222-2222-2222-2222-222222222222', '5500000000000004', 6, 2025, '456', 'Bob Johnson');

-- Subscriptions
INSERT INTO subscriptions (id, organization_id, plan_name, status) VALUES
    ('sub11111-1111-1111-1111-111111111111', '11111111-1111-1111-1111-111111111111', 'enterprise', 'active'),
    ('sub22222-2222-2222-2222-222222222222', '22222222-2222-2222-2222-222222222222', 'startup', 'active'),
    ('sub33333-3333-3333-3333-333333333333', '33333333-3333-3333-3333-333333333333', 'free', 'trialing');

-- Projects
INSERT INTO projects (id, organization_id, name, description, status) VALUES
    ('proj1111-1111-1111-1111-111111111111', '11111111-1111-1111-1111-111111111111', 'Website Redesign', 'Complete overhaul of company website', 'active'),
    ('proj2222-2222-2222-2222-222222222222', '11111111-1111-1111-1111-111111111111', 'Mobile App', 'iOS and Android app development', 'active'),
    ('proj3333-3333-3333-3333-333333333333', '22222222-2222-2222-2222-222222222222', 'MVP Launch', 'Initial product launch', 'active');

-- Tasks
INSERT INTO tasks (id, project_id, assignee_id, title, description, status, priority) VALUES
    ('task1111-1111-1111-1111-111111111111', 'proj1111-1111-1111-1111-111111111111', 'aaaa1111-1111-1111-1111-111111111111', 'Design homepage mockup', 'Create initial design concepts', 'in_progress', 1),
    ('task2222-2222-2222-2222-222222222222', 'proj1111-1111-1111-1111-111111111111', 'aaaa2222-2222-2222-2222-222222222222', 'Implement responsive layout', 'Make site mobile-friendly', 'todo', 2),
    ('task3333-3333-3333-3333-333333333333', 'proj2222-2222-2222-2222-222222222222', 'aaaa1111-1111-1111-1111-111111111111', 'Setup React Native project', 'Initialize mobile app codebase', 'done', 1),
    ('task4444-4444-4444-4444-444444444444', 'proj3333-3333-3333-3333-333333333333', 'bbbb1111-1111-1111-1111-111111111111', 'Deploy to production', 'Launch MVP to production servers', 'review', 1);

-- Invoices
INSERT INTO invoices (id, organization_id, amount_cents, currency, status, due_date) VALUES
    ('inv11111-1111-1111-1111-111111111111', '11111111-1111-1111-1111-111111111111', 99900, 'USD', 'paid', '2024-01-15'),
    ('inv22222-2222-2222-2222-222222222222', '11111111-1111-1111-1111-111111111111', 99900, 'USD', 'sent', '2024-02-15'),
    ('inv33333-3333-3333-3333-333333333333', '22222222-2222-2222-2222-222222222222', 29900, 'USD', 'paid', '2024-01-20');
