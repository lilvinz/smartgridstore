FROM postgres:11.0

ADD custom.sh /docker-entrypoint-initdb.d/
ADD entrypoint.sh /bin

ENTRYPOINT ["/bin/entrypoint.sh"]
CMD ["postgres"]
