# {{ type_name }}

Congratulations on starting development! Next steps:

1. Write the JSON schema describing your resource, `{{ schema_path.name }}`
2. The RPDK will automatically generate the correct resource model from the
   schema whenever the project is built via Make. You can also do this manually
   with the following command: `{{ executable }} generate`
3. Implement your resource handlers by adding code to provision your resources in the various Handler classes.


Please don't modify files `{{ files }}`, as will be
automatically overwritten.
