# Example::S3::Bucket BucketEncryption

## Syntax

To declare this entity in your AWS CloudFormation template, use the following syntax:

### JSON

<pre>
{
    "<a href="#serversideencryptionconfiguration" title="ServerSideEncryptionConfiguration">ServerSideEncryptionConfiguration</a>" : <i>[ <a href="serversideencryptionrule.md">ServerSideEncryptionRule</a>, ... ]</i>
}
</pre>

### YAML

<pre>
<a href="#serversideencryptionconfiguration" title="ServerSideEncryptionConfiguration">ServerSideEncryptionConfiguration</a>: <i>
      - <a href="serversideencryptionrule.md">ServerSideEncryptionRule</a></i>
</pre>

## Properties

#### ServerSideEncryptionConfiguration

_Required_: Yes

_Type_: List of <a href="serversideencryptionrule.md">ServerSideEncryptionRule</a>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

