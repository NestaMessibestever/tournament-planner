-- Tournament Planner MySQL Database Initialization Script
-- This script creates all the tables and relationships needed for the application

-- Create database and user
CREATE DATABASE IF NOT EXISTS tournament_planner CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Create a dedicated user for the application (more secure than using root)
CREATE USER IF NOT EXISTS 'tournament_user'@'localhost' IDENTIFIED BY 'tournament_pass_2024';
GRANT ALL PRIVILEGES ON tournament_planner.* TO 'tournament_user'@'localhost';
FLUSH PRIVILEGES;

USE tournament_planner;

-- Users table: Stores all system users (organizers, participants, admins)
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(36) PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(100) NOT NULL,
    phone VARCHAR(20),
    role ENUM('user', 'organizer', 'admin') DEFAULT 'user',
    email_verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_email (email),
    INDEX idx_role (role)
) ENGINE=InnoDB;

-- Sports table: Predefined sports with their specific rules
CREATE TABLE IF NOT EXISTS sports (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    scoring_type VARCHAR(50),
    rules JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB;

-- Tournaments table: Core tournament information with constraint fields
CREATE TABLE IF NOT EXISTS tournaments (
    id VARCHAR(36) PRIMARY KEY,
    organizer_id VARCHAR(36) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    sport_id VARCHAR(36),
    format_type ENUM('single_elimination', 'double_elimination', 'round_robin', 'swiss', 'group_to_knockout') NOT NULL,
    format_config JSON,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    timezone VARCHAR(50) NOT NULL DEFAULT 'UTC',
    -- Constraint fields (our key innovation!)
    max_matches_per_day INT NOT NULL,
    operational_hours JSON NOT NULL,
    avg_match_duration INT NOT NULL COMMENT 'in minutes',
    buffer_time INT DEFAULT 5 COMMENT 'in minutes',
    -- Registration settings
    registration_deadline TIMESTAMP NULL,
    entry_fee DECIMAL(10,2) DEFAULT 0.00,
    allow_onsite_payment BOOLEAN DEFAULT FALSE,
    -- Capacity (automatically calculated)
    capacity_limit INT NOT NULL,
    current_participants INT DEFAULT 0,
    -- Status and visibility
    status ENUM('draft', 'published', 'registration_open', 'registration_closed', 'in_progress', 'completed', 'cancelled') DEFAULT 'draft',
    is_public BOOLEAN DEFAULT FALSE,
    custom_fields JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (organizer_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (sport_id) REFERENCES sports(id) ON DELETE SET NULL,
    INDEX idx_organizer (organizer_id),
    INDEX idx_status (status),
    INDEX idx_dates (start_date, end_date)
) ENGINE=InnoDB;

-- Venues table: Physical locations where matches are played
CREATE TABLE IF NOT EXISTS venues (
    id VARCHAR(36) PRIMARY KEY,
    tournament_id VARCHAR(36) NOT NULL,
    name VARCHAR(100) NOT NULL,
    type ENUM('court', 'field', 'table', 'mat', 'custom') NOT NULL,
    availability_rules JSON,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tournament_id) REFERENCES tournaments(id) ON DELETE CASCADE,
    INDEX idx_tournament (tournament_id)
) ENGINE=InnoDB;

-- Participants table: Players or teams in the system
CREATE TABLE IF NOT EXISTS participants (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36),
    name VARCHAR(255) NOT NULL,
    type ENUM('individual', 'team') NOT NULL,
    contact_email VARCHAR(255),
    contact_phone VARCHAR(20),
    total_matches_played INT DEFAULT 0,
    total_matches_won INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    INDEX idx_user (user_id),
    INDEX idx_type (type)
) ENGINE=InnoDB;

-- Tournament Participants junction table
CREATE TABLE IF NOT EXISTS tournament_participants (
    tournament_id VARCHAR(36) NOT NULL,
    participant_id VARCHAR(36) NOT NULL,
    seed INT,
    division VARCHAR(50),
    group_name VARCHAR(50),
    payment_status ENUM('pending', 'paid', 'refunded', 'waived') DEFAULT 'pending',
    checked_in BOOLEAN DEFAULT FALSE,
    registration_data JSON,
    registered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (tournament_id, participant_id),
    FOREIGN KEY (tournament_id) REFERENCES tournaments(id) ON DELETE CASCADE,
    FOREIGN KEY (participant_id) REFERENCES participants(id) ON DELETE CASCADE,
    INDEX idx_tournament (tournament_id),
    INDEX idx_participant (participant_id)
) ENGINE=InnoDB;

-- Matches table: Individual games/matches in tournaments
CREATE TABLE IF NOT EXISTS matches (
    id VARCHAR(36) PRIMARY KEY,
    tournament_id VARCHAR(36) NOT NULL,
    round_number INT NOT NULL,
    match_number INT NOT NULL,
    stage VARCHAR(50) NOT NULL DEFAULT 'main',
    group_name VARCHAR(50),
    participant1_id VARCHAR(36),
    participant2_id VARCHAR(36),
    winner_id VARCHAR(36),
    score1 INT,
    score2 INT,
    score_details JSON,
    status ENUM('pending', 'scheduled', 'in_progress', 'completed', 'cancelled', 'postponed', 'walkover') DEFAULT 'pending',
    scheduled_datetime TIMESTAMP NULL,
    actual_start_time TIMESTAMP NULL,
    actual_end_time TIMESTAMP NULL,
    venue_id VARCHAR(36),
    referee_id VARCHAR(36),
    next_match_id VARCHAR(36),
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (tournament_id) REFERENCES tournaments(id) ON DELETE CASCADE,
    FOREIGN KEY (participant1_id) REFERENCES participants(id) ON DELETE SET NULL,
    FOREIGN KEY (participant2_id) REFERENCES participants(id) ON DELETE SET NULL,
    FOREIGN KEY (winner_id) REFERENCES participants(id) ON DELETE SET NULL,
    FOREIGN KEY (venue_id) REFERENCES venues(id) ON DELETE SET NULL,
    FOREIGN KEY (next_match_id) REFERENCES matches(id) ON DELETE SET NULL,
    INDEX idx_tournament (tournament_id),
    INDEX idx_status (status),
    INDEX idx_schedule (scheduled_datetime),
    INDEX idx_venue (venue_id)
) ENGINE=InnoDB;

-- Referees table
CREATE TABLE IF NOT EXISTS referees (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36),
    organizer_id VARCHAR(36) NOT NULL,
    name VARCHAR(100) NOT NULL,
    certification_level VARCHAR(50),
    contact_email VARCHAR(255),
    contact_phone VARCHAR(20),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (organizer_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_organizer (organizer_id)
) ENGINE=InnoDB;

-- Referee assignments
CREATE TABLE IF NOT EXISTS match_referees (
    match_id VARCHAR(36) NOT NULL,
    referee_id VARCHAR(36) NOT NULL,
    role ENUM('main', 'assistant', 'observer') DEFAULT 'main',
    PRIMARY KEY (match_id, referee_id),
    FOREIGN KEY (match_id) REFERENCES matches(id) ON DELETE CASCADE,
    FOREIGN KEY (referee_id) REFERENCES referees(id) ON DELETE CASCADE
) ENGINE=InnoDB;

-- Waitlist for full tournaments
CREATE TABLE IF NOT EXISTS tournament_waitlist (
    id VARCHAR(36) PRIMARY KEY,
    tournament_id VARCHAR(36) NOT NULL,
    participant_id VARCHAR(36) NOT NULL,
    position INT NOT NULL,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tournament_id) REFERENCES tournaments(id) ON DELETE CASCADE,
    FOREIGN KEY (participant_id) REFERENCES participants(id) ON DELETE CASCADE,
    UNIQUE KEY unique_waitlist (tournament_id, participant_id),
    INDEX idx_tournament_position (tournament_id, position)
) ENGINE=InnoDB;

-- System configuration table (for flexible app config)
CREATE TABLE IF NOT EXISTS system_config (
    config_key VARCHAR(100) PRIMARY KEY,
    config_value TEXT NOT NULL,
    description TEXT,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB;

-- Insert default system configurations
INSERT INTO system_config (config_key, config_value, description) VALUES
('default_match_duration', '30', 'Default match duration in minutes'),
('default_buffer_time', '5', 'Default buffer time between matches in minutes'),
('max_tournaments_per_organizer', '10', 'Maximum active tournaments per organizer'),
('enable_waitlist', 'true', 'Enable waitlist feature for full tournaments'),
('enable_referee_management', 'true', 'Enable referee assignment features'),
('email_notifications_enabled', 'false', 'Enable email notifications (requires SendGrid)'),
('maintenance_mode', 'false', 'Put system in maintenance mode');

-- Insert some default sports
INSERT INTO sports (id, name, scoring_type, rules) VALUES
('sport-tennis', 'Tennis', 'sets', '{"sets": 3, "games_per_set": 6, "tiebreak": true}'),
('sport-soccer', 'Soccer', 'goals', '{"halves": 2, "minutes_per_half": 45}'),
('sport-basketball', 'Basketball', 'points', '{"quarters": 4, "minutes_per_quarter": 10}'),
('sport-volleyball', 'Volleyball', 'sets', '{"sets": 5, "points_per_set": 25, "final_set_points": 15}'),
('sport-badminton', 'Badminton', 'games', '{"games": 3, "points_per_game": 21}'),
('sport-table-tennis', 'Table Tennis', 'games', '{"games": 5, "points_per_game": 11}');

-- Create a default admin user (password: admin123)
-- Note: This is bcrypt hash for 'admin123' - CHANGE IN PRODUCTION!
INSERT INTO users (id, email, password_hash, full_name, role, email_verified) VALUES
('admin-001', 'admin@tournament.local', '$2a$10$YKBNkLrH9Q3XJ9RqK8eQt.IYFxb2cB4F6/KjHDLbUqXnqFgLCNFfW', 'System Admin', 'admin', TRUE);