自定义控制器监控Service

...
  annotations:
    ingress/http: "true"
...

新增Service：包含指定注解，创建ingress；不包含指定注解，忽略
删除Service：Service 与 Ingress 通过OwnerReference 绑定，删除ingress
更新Service：包含指定注解，检查ingress是否存在，不存在则创建，存在则忽略；不包含指定注解，检查ingress是否存在，存在则删除，不存在则忽略