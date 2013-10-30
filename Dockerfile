FROM ubuntu
EXPOSE 80:8080
RUN apt-get install -y ca-certificates
ADD config.json /etc/subgun.conf
ADD subgun /bin/
CMD /bin/subgun /etc/subgun.conf
