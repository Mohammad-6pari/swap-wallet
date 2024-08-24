# Swap Wallet

Swap Wallet is a local cryptocurrency exchange service that allows users to manage their cryptocurrency balances and perform conversions between different cryptocurrencies.

## API Documentation

For detailed API documentation, please refer to the Postman collection:

[Swap Wallet API Documentation](https://documenter.getpostman.com/view/22369840/2sAXjF8ufH)

This documentation provides information on all available endpoints, request parameters, and responses.

## Running the Project

To run the Swap Wallet project locally using Docker, follow these steps:

1. Ensure Docker and Docker Compose are installed on your machine.

2. Clone the repository:

    ```bash
    git clone https://github.com/your-username/swap-wallet.git
    cd swap-wallet
    ```

3. Create a `.env` file in the root directory and configure the necessary environment variables:

    ```bash
    DB_USER=postgres
    DB_HOST=db
    DB_PORT=5432
    DB_PASSWORD=password
    DB_NAME=swap_wallet
    APP_PORT=8080
    JWT_SECRET=swap_wallet
    ```

4. Build and start the Docker containers:

    ```bash
    sudo docker compose up --build
    ```

5. The application will be available at `http://localhost:8080`.
