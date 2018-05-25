# go-fileweb

Web service that stores and retrieves files. Written in Go and uploaded to Amazon's ElasticBeanstalk. 

Application Routes:

* / (Root) - Upload form to add a file to the repository - Returns the ID of the file for retrieval. Response as JSON
* /Download/{ID} - Downloads the file to the client - keeps the original file name on download.
* /list - Lists the files on the system and specifies the ID, Name, & Size of each file. Response as JSON

Considerations:

* The List method is poorly optimized by making a query to S3 on each file. To resolve this I would add a concurrent data repository such as DynamoDB in order to make a single call to the resource for faster performance.
* Downloads & Uploads for very large files is not taken under considerations. 
