FROM busybox
EXPOSE 80:8080
ADD config.json /etc/subgun.conf
ADD subgun /bin/
CMD /bin/subgun /etc/subgun.conf
