# Usar la imagen base de Go
FROM golang:1.23

# Establecer el directorio de trabajo dentro del contenedor
WORKDIR /app

# Copiar los archivos del proyecto al contenedor
COPY . .

RUN rm go.mod go.sum

#RUN go mod tidy && go list -m all

# Descargar las dependencias
RUN go mod init products-api && go mod tidy

# Construir la aplicación
RUN go build -o main .

# Especificar el puerto en el que se ejecutará la aplicación
EXPOSE 8080

# Comando para ejecutar la aplicación
CMD ["/app/main"]