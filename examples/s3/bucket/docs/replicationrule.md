# Example::S3::Bucket ReplicationRule

## Syntax

To declare this entity in your AWS CloudFormation template, use the following syntax:

### JSON

<pre>
{
    "<a href="#status" title="Status">Status</a>" : <i>String</i>,
    "<a href="#destination" title="Destination">Destination</a>" : <i><a href="replicationdestination.md">ReplicationDestination</a></i>,
    "<a href="#filter" title="Filter">Filter</a>" : <i><a href="replicationrulefilter.md">ReplicationRuleFilter</a></i>,
    "<a href="#priority" title="Priority">Priority</a>" : <i>Double</i>,
    "<a href="#sourceselectioncriteria" title="SourceSelectionCriteria">SourceSelectionCriteria</a>" : <i><a href="sourceselectioncriteria.md">SourceSelectionCriteria</a></i>,
    "<a href="#id" title="Id">Id</a>" : <i>String</i>,
    "<a href="#prefix" title="Prefix">Prefix</a>" : <i>String</i>,
    "<a href="#deletemarkerreplication" title="DeleteMarkerReplication">DeleteMarkerReplication</a>" : <i><a href="deletemarkerreplication.md">DeleteMarkerReplication</a></i>
}
</pre>

### YAML

<pre>
<a href="#status" title="Status">Status</a>: <i>String</i>
<a href="#destination" title="Destination">Destination</a>: <i><a href="replicationdestination.md">ReplicationDestination</a></i>
<a href="#filter" title="Filter">Filter</a>: <i><a href="replicationrulefilter.md">ReplicationRuleFilter</a></i>
<a href="#priority" title="Priority">Priority</a>: <i>Double</i>
<a href="#sourceselectioncriteria" title="SourceSelectionCriteria">SourceSelectionCriteria</a>: <i><a href="sourceselectioncriteria.md">SourceSelectionCriteria</a></i>
<a href="#id" title="Id">Id</a>: <i>String</i>
<a href="#prefix" title="Prefix">Prefix</a>: <i>String</i>
<a href="#deletemarkerreplication" title="DeleteMarkerReplication">DeleteMarkerReplication</a>: <i><a href="deletemarkerreplication.md">DeleteMarkerReplication</a></i>
</pre>

## Properties

#### Status

_Required_: Yes

_Type_: String

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### Destination

_Required_: Yes

_Type_: <a href="replicationdestination.md">ReplicationDestination</a>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### Filter

_Required_: No

_Type_: <a href="replicationrulefilter.md">ReplicationRuleFilter</a>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### Priority

_Required_: No

_Type_: Double

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### SourceSelectionCriteria

_Required_: No

_Type_: <a href="sourceselectioncriteria.md">SourceSelectionCriteria</a>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### Id

_Required_: No

_Type_: String

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### Prefix

_Required_: No

_Type_: String

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### DeleteMarkerReplication

_Required_: No

_Type_: <a href="deletemarkerreplication.md">DeleteMarkerReplication</a>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

