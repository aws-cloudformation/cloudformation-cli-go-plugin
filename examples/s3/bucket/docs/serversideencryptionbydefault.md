# Example::S3::Bucket ServerSideEncryptionByDefault

## Syntax

To declare this entity in your AWS CloudFormation template, use the following syntax:

### JSON

<pre>
{
    "<a href="#ssealgorithm" title="SSEAlgorithm">SSEAlgorithm</a>" : <i>String</i>,
    "<a href="#kmsmasterkeyid" title="KMSMasterKeyID">KMSMasterKeyID</a>" : <i>String</i>
}
</pre>

### YAML

<pre>
<a href="#ssealgorithm" title="SSEAlgorithm">SSEAlgorithm</a>: <i>String</i>
<a href="#kmsmasterkeyid" title="KMSMasterKeyID">KMSMasterKeyID</a>: <i>String</i>
</pre>

## Properties

#### SSEAlgorithm

_Required_: Yes

_Type_: String

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### KMSMasterKeyID

_Required_: No

_Type_: String

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)
