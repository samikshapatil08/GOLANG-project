CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL
);

CREATE TABLE listings (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    price DOUBLE PRECISION NOT NULL,
    category VARCHAR(100) NOT NULL,
    seller_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE bookmarks (
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    listing_id INT NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, listing_id)
);

CREATE INDEX listings_search_idx ON listings USING gin(to_tsvector('english', title || ' ' || description));
