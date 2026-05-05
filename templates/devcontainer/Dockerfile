FROM {{.JoomlaImage}}

RUN apt-get update && apt-get install -y \
    mariadb-client \ 
    # Install Node.js (LTS version)
    && curl -fsSL https://deb.nodesource.com/setup_lts.x | bash - \
    && apt-get install -y nodejs \
    && \
    # Install Xdebug
    pecl install xdebug \
    # Clean up apt cache to reduce image size
    && apt-get clean && rm -rf /var/lib/apt/lists/*
