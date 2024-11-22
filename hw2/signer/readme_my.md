Для работы `-race` в команде
```
go test -v -race
```
нужно скачать архив MinGW-w64 for Windows с https://winlibs.com/

Настройте переменные окружения `C:\mingw64\bin`

Проверьте установку в cmd `gcc --version`

Использование с Go и CGO
Теперь MinGW настроен и может быть использован для компиляции программ с использованием CGO. Убедитесь, что CGO_ENABLED включен, и задайте MinGW в качестве компилятора:
cmd
```
set CGO_ENABLED=1
```

Перезагрузить IDE

Запустить 
```
go test -v -race
```