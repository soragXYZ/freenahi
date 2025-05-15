DROP TABLE IF EXISTS loan;
CREATE TABLE loan (
    loan_account_id INT NOT NULL,
    total_amount FLOAT NOT NULL,
    available_amount FLOAT NOT NULL,
    used_amount FLOAT NOT NULL,
    subscription_date VARCHAR(255) NOT NULL,
    maturity_date VARCHAR(255) NOT NULL,
    start_repayment_date VARCHAR(255) NOT NULL,
    is_deferred BOOLEAN NOT NULL,
    next_payment_amount FLOAT NOT NULL,
    next_payment_date VARCHAR(255) NOT NULL,
    rate FLOAT NOT NULL,
    nb_payments_left INT UNSIGNED NOT NULL,
    nb_payments_done INT UNSIGNED NOT NULL,
    nb_payments_total INT UNSIGNED NOT NULL,
    last_payment_amount FLOAT NOT NULL,
    last_payment_date VARCHAR(255) NOT NULL,
    account_label VARCHAR(255) NOT NULL,
    insurance_label VARCHAR(255) NOT NULL,
    insurance_amount FLOAT NOT NULL,
    insurance_rate FLOAT NOT NULL,
    duration INT UNSIGNED NOT NULL,
    loan_type VARCHAR(255) NOT NULL,

    PRIMARY KEY (`loan_account_id`),
    FOREIGN KEY (`loan_account_id`) REFERENCES bankAccount(`account_id`)
);
