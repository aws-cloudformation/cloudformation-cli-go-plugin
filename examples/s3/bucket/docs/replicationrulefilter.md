# Example::S3::Bucket ReplicationRuleFilter

## Syntax

To declare this entity in your AWS CloudFormation template, use the following syntax:

### JSON

<pre>
{
    "<a href="#prefix" title="Prefix">Prefix</a>" : <i>String</i>,
    "<a href="#and" title="And">And</a>" : <i><a href="replicationruleandoperator.md">ReplicationRuleAndOperator</a></i>,
    "<a href="#tagfilter" title="TagFilter">TagFilter</a>" : <i><a href="tagfilter.md">TagFilter</a></i>
}
</pre>

### YAML

<pre>
<a href="#prefix" title="Prefix">Prefix</a>: <i>String</i>
<a href="#and" title="And">And</a>: <i><a href="replicationruleandoperator.md">ReplicationRuleAndOperator</a></i>
<a href="#tagfilter" title="TagFilter">TagFilter</a>: <i><a href="tagfilter.md">TagFilter</a></i>
</pre>

## Properties

#### Prefix

_Required_: No

_Type_: String

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### And

_Required_: No

_Type_: <a href="replicationruleandoperator.md">ReplicationRuleAndOperator</a>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### TagFilter

_Required_: No

_Type_: <a href="tagfilter.md">TagFilter</a>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)
