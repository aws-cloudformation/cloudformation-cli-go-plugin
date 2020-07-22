# Example::S3::Bucket LambdaConfiguration

## Syntax

To declare this entity in your AWS CloudFormation template, use the following syntax:

### JSON

<pre>
{
    "<a href="#function" title="Function">Function</a>" : <i>String</i>,
    "<a href="#event" title="Event">Event</a>" : <i>String</i>,
    "<a href="#filter" title="Filter">Filter</a>" : <i><a href="notificationfilter.md">NotificationFilter</a></i>
}
</pre>

### YAML

<pre>
<a href="#function" title="Function">Function</a>: <i>String</i>
<a href="#event" title="Event">Event</a>: <i>String</i>
<a href="#filter" title="Filter">Filter</a>: <i><a href="notificationfilter.md">NotificationFilter</a></i>
</pre>

## Properties

#### Function

_Required_: Yes

_Type_: String

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### Event

_Required_: Yes

_Type_: String

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### Filter

_Required_: No

_Type_: <a href="notificationfilter.md">NotificationFilter</a>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

