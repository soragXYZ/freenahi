DROP TABLE IF EXISTS bankAccount, tx;
CREATE TABLE bankAccount (
    account_id INT NOT NULL,
    user_id INT NOT NULL,
    bank_original_name VARCHAR(255) NOT NULL,
    bank_number VARCHAR(255) NOT NULL,
    original_name VARCHAR(255) NOT NULL,
    balance FLOAT NOT NULL,
    last_update DATETIME NOT NULL,
    iban VARCHAR(255) NOT NULL,
    currency VARCHAR(255) NOT NULL,
    account_type VARCHAR(255) NOT NULL,
    usage_type VARCHAR(255) NOT NULL,

    PRIMARY KEY (`account_id`)
);
