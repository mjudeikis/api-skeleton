FROM scratch
ADD /api /
CMD ["/api"]
EXPOSE 8000
ENV DB_PATH /persistance/database.db
VOLUME /persistance 
ADD ca-certificates.crt /etc/ssl/certs/