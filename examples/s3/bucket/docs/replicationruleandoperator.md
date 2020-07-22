# Example::S3::Bucket ReplicationRuleAndOperator

## Syntax

To declare this entity in your AWS CloudFormation template, use the following syntax:

### JSON

<pre>
{
    "<a href="#tagfilters" title="TagFilters">TagFilters</a>" : <i>[ <a href="tagfilter.md">TagFilter</a>, ... ]</i>,
    "<a href="#prefix" title="Prefix">Prefix</a>" : <i>String</i>
}
</pre>

### YAML

<pre>
<a href="#tagfilters" title="TagFilters">TagFilters</a>: <i>
      - <a href="tagfilter.md">TagFilter</a></i>
<a href="#prefix" title="Prefix">Prefix</a>: <i>String</i>
</pre>

## Properties

#### TagFilters

_Required_: No

_Type_: List of <a href="tagfilter.md">TagFilter</a>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### Prefix

_Required_: No

_Type_: String

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

