# Example::S3::Bucket NotificationFilter S3Key

## Syntax

To declare this entity in your AWS CloudFormation template, use the following syntax:

### JSON

<pre>
{
    "<a href="#rules" title="Rules">Rules</a>" : <i>[ <a href="filterrule.md">FilterRule</a>, ... ]</i>
}
</pre>

### YAML

<pre>
<a href="#rules" title="Rules">Rules</a>: <i>
      - <a href="filterrule.md">FilterRule</a></i>
</pre>

## Properties

#### Rules

_Required_: Yes

_Type_: List of <a href="filterrule.md">FilterRule</a>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)
