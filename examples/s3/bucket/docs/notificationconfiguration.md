# Example::S3::Bucket NotificationConfiguration

## Syntax

To declare this entity in your AWS CloudFormation template, use the following syntax:

### JSON

<pre>
{
    "<a href="#queueconfigurations" title="QueueConfigurations">QueueConfigurations</a>" : <i>[ <a href="queueconfiguration.md">QueueConfiguration</a>, ... ]</i>,
    "<a href="#lambdaconfigurations" title="LambdaConfigurations">LambdaConfigurations</a>" : <i>[ <a href="lambdaconfiguration.md">LambdaConfiguration</a>, ... ]</i>,
    "<a href="#topicconfigurations" title="TopicConfigurations">TopicConfigurations</a>" : <i>[ <a href="topicconfiguration.md">TopicConfiguration</a>, ... ]</i>
}
</pre>

### YAML

<pre>
<a href="#queueconfigurations" title="QueueConfigurations">QueueConfigurations</a>: <i>
      - <a href="queueconfiguration.md">QueueConfiguration</a></i>
<a href="#lambdaconfigurations" title="LambdaConfigurations">LambdaConfigurations</a>: <i>
      - <a href="lambdaconfiguration.md">LambdaConfiguration</a></i>
<a href="#topicconfigurations" title="TopicConfigurations">TopicConfigurations</a>: <i>
      - <a href="topicconfiguration.md">TopicConfiguration</a></i>
</pre>

## Properties

#### QueueConfigurations

_Required_: No

_Type_: List of <a href="queueconfiguration.md">QueueConfiguration</a>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### LambdaConfigurations

_Required_: No

_Type_: List of <a href="lambdaconfiguration.md">LambdaConfiguration</a>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### TopicConfigurations

_Required_: No

_Type_: List of <a href="topicconfiguration.md">TopicConfiguration</a>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)
