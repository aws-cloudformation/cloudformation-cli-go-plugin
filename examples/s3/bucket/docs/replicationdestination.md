# Example::S3::Bucket ReplicationDestination

## Syntax

To declare this entity in your AWS CloudFormation template, use the following syntax:

### JSON

<pre>
{
    "<a href="#accesscontroltranslation" title="AccessControlTranslation">AccessControlTranslation</a>" : <i><a href="accesscontroltranslation.md">AccessControlTranslation</a></i>,
    "<a href="#account" title="Account">Account</a>" : <i>String</i>,
    "<a href="#metrics" title="Metrics">Metrics</a>" : <i><a href="metrics.md">Metrics</a></i>,
    "<a href="#bucket" title="Bucket">Bucket</a>" : <i>String</i>,
    "<a href="#encryptionconfiguration" title="EncryptionConfiguration">EncryptionConfiguration</a>" : <i><a href="encryptionconfiguration.md">EncryptionConfiguration</a></i>,
    "<a href="#storageclass" title="StorageClass">StorageClass</a>" : <i>String</i>,
    "<a href="#replicationtime" title="ReplicationTime">ReplicationTime</a>" : <i><a href="replicationtime.md">ReplicationTime</a></i>
}
</pre>

### YAML

<pre>
<a href="#accesscontroltranslation" title="AccessControlTranslation">AccessControlTranslation</a>: <i><a href="accesscontroltranslation.md">AccessControlTranslation</a></i>
<a href="#account" title="Account">Account</a>: <i>String</i>
<a href="#metrics" title="Metrics">Metrics</a>: <i><a href="metrics.md">Metrics</a></i>
<a href="#bucket" title="Bucket">Bucket</a>: <i>String</i>
<a href="#encryptionconfiguration" title="EncryptionConfiguration">EncryptionConfiguration</a>: <i><a href="encryptionconfiguration.md">EncryptionConfiguration</a></i>
<a href="#storageclass" title="StorageClass">StorageClass</a>: <i>String</i>
<a href="#replicationtime" title="ReplicationTime">ReplicationTime</a>: <i><a href="replicationtime.md">ReplicationTime</a></i>
</pre>

## Properties

#### AccessControlTranslation

_Required_: No

_Type_: <a href="accesscontroltranslation.md">AccessControlTranslation</a>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### Account

_Required_: No

_Type_: String

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### Metrics

_Required_: No

_Type_: <a href="metrics.md">Metrics</a>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### Bucket

_Required_: Yes

_Type_: String

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### EncryptionConfiguration

_Required_: No

_Type_: <a href="encryptionconfiguration.md">EncryptionConfiguration</a>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### StorageClass

_Required_: No

_Type_: String

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### ReplicationTime

_Required_: No

_Type_: <a href="replicationtime.md">ReplicationTime</a>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

