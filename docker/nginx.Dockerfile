FROM nginx:1.20.1

# Обновление пакетов
RUN apt-get -y update && \
    rm -rf /var/lib/apt/lists/*

# Переименование default.conf, чтобы избежать конфликта
RUN mv /etc/nginx/conf.d/default.conf /etc/nginx/conf.d/default.conf.bak || true

# Копируем конфигурационные файлы
COPY ./nginx_conf/nginx.conf /etc/nginx/nginx.conf
COPY ./nginx_conf/sites-enabled /etc/nginx/conf.d/

# Добавляем конфигурацию для работы в фоновом режиме
RUN echo "daemon off;" >> /etc/nginx/nginx.conf

# Заменяем содержимое index.html
RUN sed -i "0,/nginx/s/nginx/docker-nginx/i" /usr/share/nginx/html/index.html

# Создаем пользователя docker
RUN useradd -ms /bin/bash docker

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]