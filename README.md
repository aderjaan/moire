# moire
Media server to save and serve assets

## Prerequisites

* Amazon AWS access (S3 and SNS)
* MongoDB
* Go
* ImageMagick, optipng, ffmeg 

On Mac OS, packages are available on homebrew, just run `brew install mongodb go imagemagick optipng ffmpeg`

## Setup
see the [go documentation](http://golang.org/doc/code.html) to setup a proper coding structure

```
mkdir $HOME/go
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
export GO15VENDOREXPERIMENT="1"
cd $GOPATH
go get github.com/bulletind/moire
```

homebrew users may need `export GOROOT="/usr/local/opt/go/libexec"` 

Note about packaging: 
Moire contains all dependencies in the vendor folder. The vendor folder is maintained using `godep`. Vendor packages are committed in the repo itself: this way, we can have reproducible builds which continue to work even if the source repository is not available anymore.

## Install and run

Basic knowledge of Amazon AWS is required. It's a security best practice to generate IAM keys and policies to allow restrict access to the actions needed.

### Initial Configuration

add a configuration file, e.g. `$GOPATH/bin/moire.ini` with at least the following configuration

```
[S3]
AccessKey = ***
SecretKey = ***
Bucket = moire-gallery-name
```

### Installation

* Make sure mongo is running
* `cd src/github.com/bulletind/moire`
* `go install && moire -config $GOPATH/bin/moire.ini`


### Uploading your first asset and further configuration

Make a POST request to `http://127.0.0.1:8811/assets/` with the following json data:

```
{
    "mime_type": "image/jpeg",
    "name": "filename.jpg"
}
```

Moire will return a response like this:

```
{
  "data": {
    "_id": "55a78555b1b11b11de000001",
    "upload_url": "https://moire-gallery-local.s3.amazonaws.com/original_file/55a78555b1b11b11de000001/TiEmRuFUij/filename.jpg?AWSAccessKeyId=****************&Expires=1468578005&Signature=****************",
    "url": "/assets/55a78555b1b11b11de000001"
  },
  "message": "",
  "status": 200
}
```

Now, you can use the `upload_url` to directly upload the file to S3. This can be done using e.g. curl:

```
curl -vv -X PUT -T filename.jpg -H "Content-Type: image/jpeg" \ 
"https://moire-gallery-local.s3.amazonaws.com/original_file/55a78555b1b11b11de000001/TiEmRuFUij/filename.jpg?AWSAccessKeyId=****************&Expires=1468578005&Signature=****************"
```

After uploading the file to S3, moire relies on a web hook to be called in order to update the asset as being available. Next to that, it also performs background tasks to generate thumbnails for images and videos. Amazon S3 allows events to be triggered after a file has been uploaded.

Create an Amazon SNS Topic (to be triggered). In this example, `moire-gallery-upload` is used as the topic name. Change your topic policy to allow the Amazon S3 bucket to publish:

```
{
  "Version": "2008-10-17",
  "Id": "__default_policy_ID",
  "Statement": [
    {
      "Sid": "__default_statement_ID",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "SNS:Publish",
      "Resource": "arn:aws:sns:eu-west-1:***:moire-gallery-upload",
      "Condition": {
        "ArnLike": {
          "aws:SourceArn": "arn:aws:s3:*:*:moire-gallery"
        }
      }
    }
  ]
}
```

Configure your S3 bucket (select bucket > properties > events) and select ObjectCreated(all) with the SNS topic name. After uploading your next file, a message will be in your SNS Topic. You can easily test that to create an email subscription in the SNS topic. For more information about this, see the [AWS documentation(http://docs.aws.amazon.com/AmazonS3/latest/UG/SettingBucketNotifications.html)].

Now it's time to configure the web hook. As you are probably behind a firewall and/or using NAT, your local moire instance wont be accessible publicly. To work around that, you can use a tunnel service like [ngrok(https://ngrok.com)], then run `ngrok http localhost:8811`. Add a HTTPS webhook to the following endpoint: `https://host/notify/sns`

At this time of writing, there is very minimal support to handle SNS subsciptions:

1. Add an HTTPS Subsciption in the SNS Console to the following endpoint: `https://host/notify/sns`
2. SNS sends a `SubscriptionConfirmation` request
3. Moire simply panics and sends an exception email
4. Confirm the Subscription in the SNS Console using the `SubscribeURL` in the email subject
5. You are all set: SNS should have assigned a Subscription ID
