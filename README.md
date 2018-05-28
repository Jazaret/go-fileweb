# go-fileweb [![Build Status](https://travis-ci.org/Jazaret/go-fileweb.svg?branch=master)](https://travis-ci.org/Jazaret/go-fileweb)

Web service that stores and retrieves files. Written in Go and uploaded to Amazon's ElasticBeanstalk. Implements a full browser website as well as an additional endpoint that lists the files as requested. 

How to Compile/Run: 
* git clone https://github.com/Jazaret/go-fileweb
* cd go-fileweb
* go get -t -v ./...
* go build
* run go-build.exe

Endpoint Routes:

* /upload - POST form data with file object to add a file to the repository - Returns the ID of the file for retrieval. Response as JSON
* /api/download/{ID} - Endpoint that downloads the file to the client - keeps the original file name on download.
* /api/list - Endpoint that returns a list of all files in the system, their identifier, original filename, and the byte size of the file.. Response as JSON

Full website: 

http://gofilewebapp-env.hte7s2zj5y.us-east-1.elasticbeanstalk.com

Considerations:
* Code is stored on github with Travis CI for code build & tests validation. https://github.com/Jazaret/go-fileweb
* Files are stored in an S3 bucket with the original file name, content type, & size stored in the object's metadata. 
* Website uses a model / view / controller structure with HTML templates
* The list method is poorly optimized by making a query to S3 on each file to get the original file name stored in tag/metadata. To resolve this I would add a parallel data repository/database in order to make a single call to the resource for faster performance.
* There is logic for a token with Expiration on branch https://github.com/Jazaret/go-fileweb/tree/AccessToken-With-Expiration. Implementation was not merged due to time. Has hardcoded expiry time of 7 days. Also would have liked to implement access using both file ID and file access token ID, but that was not part of specifications. 
* There is no consideration as to uploading a duplicate file. We can check the file name as well as a checksum or hash stored on a seperate queriable repository to ensure that users are not uploading the same file repeatedly. 
* Downloads & Uploads for very large files is not taken under considerations. 
