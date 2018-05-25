# go-fileweb

Web service that stores and retrieves files. Written in Go and uploaded to Amazon's ElasticBeanstalk. 

Application Routes:

* / (Root) - Upload form to add a file to the repository - Returns the ID of the file for retrieval. Response as JSON
* /Download/{ID} - Downloads the file to the client - keeps the original file name on download.
* /list - Lists the files on the system and specifies the ID, Name, & Size of each file. Response as JSON

Considerations:
* Files are stored in an S3 bucket with the original file name, content type, & size stored in the object's metadata. 
* The list method is poorly optimized by making a query to S3 on each file. To resolve this I would add a parallel data repository/database in order to make a single call to the resource for faster performance.
* There is no consideration as to uploading a duplicate file. We can check the file name as well as a checksum or hash stored on a seperate queriable repository to ensure that users are not uploading the same file repeatedly. 
* Downloads & Uploads for very large files is not taken under considerations. 
