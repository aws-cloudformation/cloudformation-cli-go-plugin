# Example::S3::Bucket

Resource Type definition for Example::S3::Bucket

## Syntax

To declare this entity in your AWS CloudFormation template, use the following syntax:

### JSON

<pre>
{
    "Type" : "Example::S3::Bucket",
    "Properties" : {
        "<a href="#inventoryconfigurations" title="InventoryConfigurations">InventoryConfigurations</a>" : <i>[ <a href="inventoryconfiguration.md">InventoryConfiguration</a>, ... ]</i>,
        "<a href="#bucketencryption" title="BucketEncryption">BucketEncryption</a>" : <i><a href="bucketencryption.md">BucketEncryption</a></i>,
        "<a href="#notificationconfiguration" title="NotificationConfiguration">NotificationConfiguration</a>" : <i><a href="notificationconfiguration.md">NotificationConfiguration</a></i>,
        "<a href="#websiteconfiguration" title="WebsiteConfiguration">WebsiteConfiguration</a>" : <i><a href="websiteconfiguration.md">WebsiteConfiguration</a></i>,
        "<a href="#lifecycleconfiguration" title="LifecycleConfiguration">LifecycleConfiguration</a>" : <i><a href="lifecycleconfiguration.md">LifecycleConfiguration</a></i>,
        "<a href="#versioningconfiguration" title="VersioningConfiguration">VersioningConfiguration</a>" : <i><a href="versioningconfiguration.md">VersioningConfiguration</a></i>,
        "<a href="#metricsconfigurations" title="MetricsConfigurations">MetricsConfigurations</a>" : <i>[ <a href="metricsconfiguration.md">MetricsConfiguration</a>, ... ]</i>,
        "<a href="#accesscontrol" title="AccessControl">AccessControl</a>" : <i>String</i>,
        "<a href="#analyticsconfigurations" title="AnalyticsConfigurations">AnalyticsConfigurations</a>" : <i>[ <a href="analyticsconfiguration.md">AnalyticsConfiguration</a>, ... ]</i>,
        "<a href="#accelerateconfiguration" title="AccelerateConfiguration">AccelerateConfiguration</a>" : <i><a href="accelerateconfiguration.md">AccelerateConfiguration</a></i>,
        "<a href="#publicaccessblockconfiguration" title="PublicAccessBlockConfiguration">PublicAccessBlockConfiguration</a>" : <i><a href="publicaccessblockconfiguration.md">PublicAccessBlockConfiguration</a></i>,
        "<a href="#bucketname" title="BucketName">BucketName</a>" : <i>String</i>,
        "<a href="#corsconfiguration" title="CorsConfiguration">CorsConfiguration</a>" : <i><a href="corsconfiguration.md">CorsConfiguration</a></i>,
        "<a href="#objectlockconfiguration" title="ObjectLockConfiguration">ObjectLockConfiguration</a>" : <i><a href="objectlockconfiguration.md">ObjectLockConfiguration</a></i>,
        "<a href="#objectlockenabled" title="ObjectLockEnabled">ObjectLockEnabled</a>" : <i>Boolean</i>,
        "<a href="#loggingconfiguration" title="LoggingConfiguration">LoggingConfiguration</a>" : <i><a href="loggingconfiguration.md">LoggingConfiguration</a></i>,
        "<a href="#replicationconfiguration" title="ReplicationConfiguration">ReplicationConfiguration</a>" : <i><a href="replicationconfiguration.md">ReplicationConfiguration</a></i>,
        "<a href="#tags" title="Tags">Tags</a>" : <i>[ <a href="tag.md">Tag</a>, ... ]</i>
    }
}
</pre>

### YAML

<pre>
Type: Example::S3::Bucket
Properties:
    <a href="#inventoryconfigurations" title="InventoryConfigurations">InventoryConfigurations</a>: <i>
      - <a href="inventoryconfiguration.md">InventoryConfiguration</a></i>
    <a href="#bucketencryption" title="BucketEncryption">BucketEncryption</a>: <i><a href="bucketencryption.md">BucketEncryption</a></i>
    <a href="#notificationconfiguration" title="NotificationConfiguration">NotificationConfiguration</a>: <i><a href="notificationconfiguration.md">NotificationConfiguration</a></i>
    <a href="#websiteconfiguration" title="WebsiteConfiguration">WebsiteConfiguration</a>: <i><a href="websiteconfiguration.md">WebsiteConfiguration</a></i>
    <a href="#lifecycleconfiguration" title="LifecycleConfiguration">LifecycleConfiguration</a>: <i><a href="lifecycleconfiguration.md">LifecycleConfiguration</a></i>
    <a href="#versioningconfiguration" title="VersioningConfiguration">VersioningConfiguration</a>: <i><a href="versioningconfiguration.md">VersioningConfiguration</a></i>
    <a href="#metricsconfigurations" title="MetricsConfigurations">MetricsConfigurations</a>: <i>
      - <a href="metricsconfiguration.md">MetricsConfiguration</a></i>
    <a href="#accesscontrol" title="AccessControl">AccessControl</a>: <i>String</i>
    <a href="#analyticsconfigurations" title="AnalyticsConfigurations">AnalyticsConfigurations</a>: <i>
      - <a href="analyticsconfiguration.md">AnalyticsConfiguration</a></i>
    <a href="#accelerateconfiguration" title="AccelerateConfiguration">AccelerateConfiguration</a>: <i><a href="accelerateconfiguration.md">AccelerateConfiguration</a></i>
    <a href="#publicaccessblockconfiguration" title="PublicAccessBlockConfiguration">PublicAccessBlockConfiguration</a>: <i><a href="publicaccessblockconfiguration.md">PublicAccessBlockConfiguration</a></i>
    <a href="#bucketname" title="BucketName">BucketName</a>: <i>String</i>
    <a href="#corsconfiguration" title="CorsConfiguration">CorsConfiguration</a>: <i><a href="corsconfiguration.md">CorsConfiguration</a></i>
    <a href="#objectlockconfiguration" title="ObjectLockConfiguration">ObjectLockConfiguration</a>: <i><a href="objectlockconfiguration.md">ObjectLockConfiguration</a></i>
    <a href="#objectlockenabled" title="ObjectLockEnabled">ObjectLockEnabled</a>: <i>Boolean</i>
    <a href="#loggingconfiguration" title="LoggingConfiguration">LoggingConfiguration</a>: <i><a href="loggingconfiguration.md">LoggingConfiguration</a></i>
    <a href="#replicationconfiguration" title="ReplicationConfiguration">ReplicationConfiguration</a>: <i><a href="replicationconfiguration.md">ReplicationConfiguration</a></i>
    <a href="#tags" title="Tags">Tags</a>: <i>
      - <a href="tag.md">Tag</a></i>
</pre>

## Properties

#### InventoryConfigurations

_Required_: No

_Type_: List of <a href="inventoryconfiguration.md">InventoryConfiguration</a>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### BucketEncryption

_Required_: No

_Type_: <a href="bucketencryption.md">BucketEncryption</a>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### NotificationConfiguration

_Required_: No

_Type_: <a href="notificationconfiguration.md">NotificationConfiguration</a>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### WebsiteConfiguration

_Required_: No

_Type_: <a href="websiteconfiguration.md">WebsiteConfiguration</a>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### LifecycleConfiguration

_Required_: No

_Type_: <a href="lifecycleconfiguration.md">LifecycleConfiguration</a>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### VersioningConfiguration

_Required_: No

_Type_: <a href="versioningconfiguration.md">VersioningConfiguration</a>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### MetricsConfigurations

_Required_: No

_Type_: List of <a href="metricsconfiguration.md">MetricsConfiguration</a>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### AccessControl

_Required_: No

_Type_: String

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### AnalyticsConfigurations

_Required_: No

_Type_: List of <a href="analyticsconfiguration.md">AnalyticsConfiguration</a>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### AccelerateConfiguration

_Required_: No

_Type_: <a href="accelerateconfiguration.md">AccelerateConfiguration</a>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### PublicAccessBlockConfiguration

_Required_: No

_Type_: <a href="publicaccessblockconfiguration.md">PublicAccessBlockConfiguration</a>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### BucketName

_Required_: No

_Type_: String

_Update requires_: [Replacement](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-replacement)

#### CorsConfiguration

_Required_: No

_Type_: <a href="corsconfiguration.md">CorsConfiguration</a>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### ObjectLockConfiguration

_Required_: No

_Type_: <a href="objectlockconfiguration.md">ObjectLockConfiguration</a>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### ObjectLockEnabled

_Required_: No

_Type_: Boolean

_Update requires_: [Replacement](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-replacement)

#### LoggingConfiguration

_Required_: No

_Type_: <a href="loggingconfiguration.md">LoggingConfiguration</a>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### ReplicationConfiguration

_Required_: No

_Type_: <a href="replicationconfiguration.md">ReplicationConfiguration</a>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### Tags

_Required_: No

_Type_: List of <a href="tag.md">Tag</a>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

## Return Values

### Ref

When you pass the logical ID of this resource to the intrinsic `Ref` function, Ref returns the Id.

### Fn::GetAtt

The `Fn::GetAtt` intrinsic function returns a value for a specified attribute of this type. The following are the available attributes and sample return values.

For more information about using the `Fn::GetAtt` intrinsic function, see [Fn::GetAtt](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/intrinsic-function-reference-getatt.html).

#### RegionalDomainName

Returns the <code>RegionalDomainName</code> value.

#### Id

Returns the <code>Id</code> value.

#### WebsiteURL

Returns the <code>WebsiteURL</code> value.

#### Arn

Returns the <code>Arn</code> value.

#### DomainName

Returns the <code>DomainName</code> value.

#### DualStackDomainName

Returns the <code>DualStackDomainName</code> value.
