
# client
reads csv file and calls the service 'server' via TCP

# server
listen to the TCP port and inserts/updates on duplicate key data into the table

# remark
Ive used TCP+MySQL which is not the best for this service. 
But in the case of hard time limit i choose to use them cause I know them very well. In the production i'll better use gRPC+MongoDB. But surely it'll takes much more time

