FROM centurylink/ca-certs
ADD finalexam /
EXPOSE 2019
CMD ["/finalexam"]
