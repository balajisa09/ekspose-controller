FROM ubuntu

COPY ./ekspose-controller /usr/local/bin/ekspose-controller

ENTRYPOINT [ "/usr/local/bin/ekspose-controller" ]