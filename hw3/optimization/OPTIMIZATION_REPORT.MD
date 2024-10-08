запускаем бенчмарк 
```
m -cpuprofile="cpu.out" -memprofile="mem.out" -memprofilerate=1 .
```
запускаем профилирование по цпу
```
go tool pprof hw3.test.exe cpu.out
```
смотрим производительность FastSearch
```
list FastSearch
```

- видим замедление на
![img_1.png](img_1.png)
![img_2.png](img_2.png)
Устранение повторной компиляции паттерна поиска для всех мест, где идет поиск по регулярному выражению
![img.png](img.png)

- видим замедление на
- 
![img_3.png](img_3.png)