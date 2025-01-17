package main

import (
	"fmt"
	"math"
	"reflect"
)

func checkFloat64ToInt(val float64) bool {
	if _, fracPart := math.Modf(val); fracPart == 0 && val < float64(math.MaxInt) && val > float64(math.MinInt) {
		return true
	}
	return false
}

// parseSimpleType разбор простого типа для маппинга
//
//	targetVal это уже
func parseSimpleType(mapVal reflect.Value, targetVal reflect.Value, fieldInfo reflect.StructField) error {
	targetType := targetVal.Type().Name()
	sourceType := mapVal.Type().Name()
	// обработка int
	if targetType == "int" {
		if sourceType == "float64" {
			floatVal := mapVal.Float()
			if checkFloat64ToInt(floatVal) {
				targetVal.SetInt(int64(floatVal))
			}
		} else {
			return fmt.Errorf("invalid source for int target")
		}
	}
	// обработка string
	if targetType == "string" {
		if targetType == sourceType {
			stringVal := mapVal.String()
			targetVal.SetString(stringVal)
		} else {
			return fmt.Errorf("invalid source for string target")
		}
	}
	// обработка bool
	if targetType == "bool" {
		if targetType == sourceType {
			boolVal := mapVal.Bool()
			targetVal.SetBool(boolVal)
		} else {
			return fmt.Errorf("invalid source for bool target")
		}
	}
	return nil
}

// rec функция маппинга для рекурсии
func rec(sourceVal reflect.Value, targetVal reflect.Value) error {
	fmt.Println("-------------------------------")
	if !sourceVal.IsValid() || !targetVal.IsValid() {
		return fmt.Errorf("Invalid data")
	}
	if targetVal.Kind().String() == "struct" && sourceVal.Kind().String() != "map" {
		return fmt.Errorf("Invalid source type for data")
	}
	fmt.Println("targetVal ", targetVal, targetVal.Kind(), targetVal.Type(), targetVal.CanSet())
	fmt.Println("sourceVal ", sourceVal, sourceVal.Kind(), sourceVal.CanSet())
	// если слайс
	if targetVal.Kind() == reflect.Slice {
		// создаем буфферный массив
		newTargetVal := reflect.New(targetVal.Type())
		// проходим по слайсу источника а не таргета, т.к. там значения
		for i := 0; i < sourceVal.Len(); i++ {
			fmt.Println("slice info: ", targetVal.Type())
			//создаем новый элемент для слайса, его передаем в рекурсию для заполнения
			newTargetValElm := reflect.New(targetVal.Type().Elem())
			rec(sourceVal.Index(i).Elem(), newTargetValElm.Elem())
			// заполняем буфферный массив, т.к. reflect.Append создает новый массив и напрямую делать в массив источника нельзя
			newTargetVal = reflect.Append(targetVal, newTargetValElm.Elem())
			// кладем буфферный массив в структуру
			targetVal.Set(newTargetVal)
		}
	}
	// если структура
	if targetVal.Kind() == reflect.Struct {
		// обход структуры
		for _, fieldInfo := range reflect.VisibleFields(targetVal.Type()) {
			fmt.Println("field ", fieldInfo.Name)
			// если поле структура, то вызываем рекурсию
			if fieldInfo.Type.Kind() == reflect.Struct {
				t := targetVal.FieldByName(fieldInfo.Name)
				rec(targetVal.MapIndex(reflect.ValueOf(fieldInfo.Name)), t.Elem())
			} else { // если простой тип
				if err := parseSimpleType(sourceVal.MapIndex(reflect.ValueOf(fieldInfo.Name)).Elem(), targetVal.FieldByName(fieldInfo.Name), fieldInfo); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// i2s основная функция маппинга
func i2s(data interface{}, out interface{}) error {
	// todo
	fmt.Println("-----------------------------------new")
	if reflect.ValueOf(out).Kind() != reflect.Pointer {
		return fmt.Errorf("пришел не ссылочный тип - мы не сможем вернуть результат")
	}
	outElm := reflect.ValueOf(out).Elem()
	dataElm := reflect.ValueOf(data)
	if !dataElm.IsValid() || !outElm.IsValid() {
		return fmt.Errorf("Invalid data")
	}
	if outElm.Kind().String() == "struct" && dataElm.Kind().String() != "map" {
		return fmt.Errorf("ожидаем структуру - пришел массив")
	}
	if outElm.Kind() == reflect.Slice && outElm.Type().Elem().Kind().String() == "struct" {
		// создаем буфферный массив
		newTargetVal := reflect.New(outElm.Type())
		// проходим по слайсу источника а не таргета, т.к. там значения
		for i := 0; i < dataElm.Len(); i++ {
			fmt.Println("slice info: ", outElm.Type())
			//создаем новый элемент для слайса, его передаем в рекурсию для заполнения
			newTargetValElm := reflect.New(outElm.Type().Elem())
			if err := rec(dataElm.Index(i).Elem(), newTargetValElm.Elem()); err != nil {
				return err
			}
			// заполняем буфферный массив, т.к. reflect.Append создает новый массив и напрямую делать в массив источника нельзя
			newTargetVal = reflect.Append(outElm, newTargetValElm.Elem())
			// кладем буфферный массив в структуру
			outElm.Set(newTargetVal)
		}
	} else if outElm.Kind() == reflect.Struct {
		visibleFields := reflect.VisibleFields(outElm.Type())
		for i, fieldInfo := range visibleFields {
			if fieldInfo.Type.Kind() == reflect.Struct || fieldInfo.Type.Kind() == reflect.Slice {
				outElmField := outElm.Field(i)
				dataElmValue := dataElm.MapIndex(reflect.ValueOf(fieldInfo.Name))
				if !dataElmValue.IsValid() {
					return fmt.Errorf("поле в источнике пустое")
				}
				if err := rec(dataElmValue.Elem(), outElmField); err != nil {
					return err
				}
			} else {
				fmt.Println(fieldInfo.Type.Kind())
				dataElmValue := dataElm.MapIndex(reflect.ValueOf(fieldInfo.Name))
				if err := parseSimpleType(dataElmValue.Elem(), outElm.FieldByName(fieldInfo.Name), fieldInfo); err != nil {
					return err
				}
			}
		}

	} else {
		return fmt.Errorf("Bad data type")
	}
	return nil
}
