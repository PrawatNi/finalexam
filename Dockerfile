FROM centurylink/ca-certs
ADD main /
EXPOSE 2019
CMD ["/main"]
