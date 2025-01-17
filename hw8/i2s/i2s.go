package main

import (
	"fmt"
	"math"
	"reflect"
)

//type Simple struct {
//	ID       int
//	Username string
//	Active   bool
//}
//type Complex struct {
//	SubSimple  Simple
//	ManySimple []Simple
//	Blocks     []IDBlock
//	Test       string
//}
//type IDBlock struct {
//	ID int
//}

func checkFloat64ToInt(val float64) bool {
	if _, fracPart := math.Modf(val); fracPart == 0 && val < float64(math.MaxInt) && val > float64(math.MinInt) {
		return true
	}
	return false
}

func i2s11(data interface{}, out interface{}) error {
	// todo
	fmt.Println("-------------------------------")
	fmt.Println("out ", out)
	var outValSet, outVal reflect.Value
	outValSet = reflect.ValueOf(out).Elem()
	fmt.Println("outValSet ", outValSet, outValSet.Kind(), outValSet.Type(), outValSet.CanSet())
	dataIn := reflect.ValueOf(data)
	if outValSet.Kind() == reflect.Interface {
		//outVal = reflect.ValueOf(&out).Elem()
		outVal = outValSet.Elem() // переменная для просмотра полей структуры
		outValSet1 := outValSet   // переменная для изменения полей структуры
		fmt.Println("inVal ", outValSet1.Interface(), outValSet1.Kind(), outValSet1.Type(), outValSet1.CanSet())
		outValSet1.FieldByName("ID").SetInt(111)

	} else {
		outVal = outValSet
	}
	visibleFields := reflect.VisibleFields(outVal.Type())
	fmt.Println("visibleFields ", visibleFields)
	for i, fieldInfo := range visibleFields {
		fmt.Println("field", fieldInfo.Name)
		if fieldInfo.Type.Kind() == reflect.Struct {
			fmt.Println("outIn.Field(i)", outVal.Field(i).Type())
			n := outVal.Field(i).Interface()
			i2s(dataIn.MapIndex(reflect.ValueOf(fieldInfo.Name)), &n)
		} else {
			fmt.Println(fieldInfo.Type.Kind())
			ttt := reflect.ValueOf(outValSet).Elem()
			fmt.Println("outValSet ", ttt, ttt.Kind(), ttt.Type(), ttt.CanSet())

			ttt.FieldByName(fieldInfo.Name).SetInt(111)
			fmt.Println("element new ", outValSet.Interface())
		}
		break
	}
	fmt.Println("out!!!!!!!!!!!!!!", out)

	//fmt.Println("visibleFields ", visibleFields[0].Type.Kind())
	//outIn.FieldByName("ID").SetInt(99)
	//if visibleFields[0].Type.Kind() == reflect.Struct {
	//	fmt.Println("Struct field!!")
	//	i2s(visibleFields[0])
	//} else {
	//
	//}
	//fmt.Println("dataIn", dataIn.MapIndex(reflect.ValueOf("SubSimple")))

	return nil
}

// parseSimpleType разбор простого типа для маппинга
//
//	targetVal это уже
func parseSimpleType(mapVal reflect.Value, targetVal reflect.Value, fieldInfo reflect.StructField) error {
	fmt.Println("Change field:", fieldInfo.Name, fieldInfo.Type)
	fmt.Println("mapVal", mapVal.Kind())
	//
	//if mapVal.Elem().Kind() == reflect.Map {
	//	mapVal = mapVal.Elem().MapIndex(reflect.ValueOf(fieldInfo.Name))
	//	fmt.Println("mapVal ", mapVal.Elem().Type(), mapVal.Elem().Type().Name(), targetVal.FieldByName(fieldInfo.Name).Type().Name())
	//}

	targetType := targetVal.Type().Name()
	sourceType := mapVal.Type().Name()
	fmt.Println("targetType, sourceType", targetType, sourceType, mapVal)
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
				n := reflect.New(t.Type()).Elem()
				n.Set(t)
				fmt.Println("n ", n, n.Kind(), n.Type(), n.CanSet())
				rec(targetVal.MapIndex(reflect.ValueOf(fieldInfo.Name)), n)
			} else { // если простой тип
				//mapVal.Elem().MapIndex(reflect.ValueOf(fieldInfo.Name))
				fmt.Println("sourceVal", sourceVal, sourceVal.Type(), sourceVal.Kind())
				//dataElm.MapIndex(reflect.ValueOf(fieldInfo.Name)
				if err := parseSimpleType(sourceVal.MapIndex(reflect.ValueOf(fieldInfo.Name)).Elem(), targetVal.FieldByName(fieldInfo.Name), fieldInfo); err != nil {
					return err
				}
			}
		}
	}
	fmt.Println("targetVal.Interface()", targetVal.Interface())

	return nil
}

func i2s(data interface{}, out interface{}) error {
	// todo
	fmt.Println("-----------------------------------new")
	if reflect.ValueOf(out).Kind() != reflect.Pointer {
		return fmt.Errorf("пришел не ссылочный тип - мы не сможем вернуть результат")
	}
	outElm := reflect.ValueOf(out).Elem()
	dataElm := reflect.ValueOf(data)
	fmt.Println("outElm ", outElm, outElm.Kind(), outElm.Type())
	fmt.Println("dataIn ", dataElm, dataElm.Kind(), dataElm.Type())
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
			fmt.Println("Field:", fieldInfo.Name)
			if fieldInfo.Type.Kind() == reflect.Struct || fieldInfo.Type.Kind() == reflect.Slice {
				outElmField := outElm.Field(i)
				dataElmValue := dataElm.MapIndex(reflect.ValueOf(fieldInfo.Name))
				fmt.Println("dataElmValue", dataElmValue, dataElm)
				if !dataElmValue.IsValid() {
					return fmt.Errorf("поле в источнике пустое")
				}
				if err := rec(dataElmValue.Elem(), outElmField); err != nil {
					return err
				}
				//} else if fieldInfo.Type.Kind() == reflect.Slice {
				//	outElmField := outElm.Field(i)
				//	dataElmValue := dataElm.MapIndex(reflect.ValueOf(fieldInfo.Name))
				//	rec(dataElmValue, outElmField)
			} else {
				fmt.Println(fieldInfo.Type.Kind())
				dataElmValue := dataElm.MapIndex(reflect.ValueOf(fieldInfo.Name))
				fmt.Println("dataElm", dataElm)
				fmt.Println("dataElmValue", dataElmValue.Elem().Type().String())

				if err := parseSimpleType(dataElmValue.Elem(), outElm.FieldByName(fieldInfo.Name), fieldInfo); err != nil {
					return err
				}
			}
		}

	} else {
		return fmt.Errorf("Bad data type")
	}
	fmt.Println("out", out)

	return nil
}

//func main() {
//	smpl := Simple{
//		ID:       42,
//		Username: "rvasily",
//		Active:   true,
//	}
//	expected := []Simple{smpl, smpl}
//
//	jsonRaw, _ := json.Marshal(expected)
//
//	var tmpData interface{}
//	json.Unmarshal(jsonRaw, &tmpData)
//
//	result := []Simple{}
//	err := i2s(tmpData, &result)
//
//	if err != nil {
//		fmt.Printf("unexpected error: %v", err)
//	}
//	if !reflect.DeepEqual(expected, result) {
//		fmt.Printf("results not match\nGot:\n%#v\nExpected:\n%#v", result, expected)
//	}
//}
