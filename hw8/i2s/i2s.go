package main

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
)

type Simple struct {
	ID       int
	Username string
	Active   bool
}
type Complex struct {
	SubSimple  Simple
	ManySimple []Simple
	Blocks     []IDBlock
	Test       string
}
type IDBlock struct {
	ID int
}

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

func parseSimpleType(mapVal reflect.Value, targetVal reflect.Value, fieldInfo reflect.StructField) error {
	fmt.Println("Change field:", fieldInfo.Name, fieldInfo.Type)
	fmt.Println("mapVal", mapVal)
	if mapVal.Elem().Kind() == reflect.Map {
		mapVal = mapVal.Elem().MapIndex(reflect.ValueOf(fieldInfo.Name))
		fmt.Println("mapVal ", mapVal.Elem().Type(), mapVal.Elem().Type().Name(), targetVal.FieldByName(fieldInfo.Name).Type().Name())
	}

	targetType := targetVal.FieldByName(fieldInfo.Name).Type().Name()
	sourceType := mapVal.Elem().Type().Name()
	// обработка int
	if targetType == "int" {
		floatVal := mapVal.Elem().Float()
		if sourceType == "float64" && checkFloat64ToInt(floatVal) {
			targetVal.FieldByName(fieldInfo.Name).SetInt(int64(floatVal))
		} else {
			return fmt.Errorf("invalid source for int target")
		}
	}
	// обработка string
	if targetType == "string" {
		stringVal := mapVal.Elem().String()
		if sourceType == "string" {
			targetVal.FieldByName(fieldInfo.Name).SetString(stringVal)
		} else {
			return fmt.Errorf("invalid source for string target")
		}
	}

	// обработка bool
	if targetType == "bool" {
		boolVal := mapVal.Elem().Bool()
		if sourceType == "bool" {
			targetVal.FieldByName(fieldInfo.Name).SetBool(boolVal)
		} else {
			return fmt.Errorf("invalid source for bool target")
		}
	}
	return nil
}

func rec(sourceVal reflect.Value, targetVal reflect.Value) error {
	fmt.Println("-------------------------------")
	fmt.Println("targetVal ", targetVal, targetVal.Kind(), targetVal.Type(), targetVal.CanSet())
	fmt.Println("sourceVal ", sourceVal, sourceVal.Kind(), sourceVal.Type(), sourceVal.CanSet(), sourceVal.Elem().Len())
	// если слайс
	if targetVal.Kind() == reflect.Slice {
		// создаем буфферный массив
		newTargetVal := reflect.New(targetVal.Type())
		// проходим по слайсу источника а не таргета, т.к. там значения
		for i := 0; i < sourceVal.Elem().Len(); i++ {
			fmt.Println("slice info: ", targetVal.Type())
			//создаем новый элемент для слайса, его передаем в рекурсию для заполнения
			newTargetValElm := reflect.New(targetVal.Type().Elem())
			rec(sourceVal.Elem().Index(i), newTargetValElm.Elem())
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
				if err := parseSimpleType(sourceVal, targetVal, fieldInfo); err != nil {
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
	outElm := reflect.ValueOf(out).Elem()
	dataElm := reflect.ValueOf(data)
	fmt.Println("outElm ", outElm, outElm.Kind(), outElm.Type())
	fmt.Println("dataIn ", dataElm, dataElm.Kind(), dataElm.Type())

	visibleFields := reflect.VisibleFields(outElm.Type())

	for i, fieldInfo := range visibleFields {
		fmt.Println("Field:", fieldInfo.Name)
		if fieldInfo.Type.Kind() == reflect.Struct || fieldInfo.Type.Kind() == reflect.Slice {
			outElmField := outElm.Field(i)
			dataElmValue := dataElm.MapIndex(reflect.ValueOf(fieldInfo.Name))
			fmt.Println("dataElmValue", dataElmValue)
			rec(dataElmValue, outElmField)
			//} else if fieldInfo.Type.Kind() == reflect.Slice {
			//	outElmField := outElm.Field(i)
			//	dataElmValue := dataElm.MapIndex(reflect.ValueOf(fieldInfo.Name))
			//	rec(dataElmValue, outElmField)
		} else {
			fmt.Println(fieldInfo.Type.Kind())
			dataElmValue := dataElm.MapIndex(reflect.ValueOf(fieldInfo.Name))
			fmt.Println("dataElm", dataElm)
			fmt.Println("dataElmValue", dataElmValue.Elem().Type().String())
			parseSimpleType(dataElmValue, outElm, fieldInfo)
		}
	}
	fmt.Println("out", out)

	return nil
}

func main() {
	smpl := Simple{
		ID:       42,
		Username: "rvasily",
		Active:   true,
	}
	expected := &Complex{
		SubSimple:  smpl,
		ManySimple: []Simple{smpl, smpl},
		Blocks:     []IDBlock{IDBlock{42}, IDBlock{42}},
		Test:       "test",
	}

	jsonRaw, _ := json.Marshal(expected)
	// fmt.Println(string(jsonRaw))

	var tmpData interface{}
	json.Unmarshal(jsonRaw, &tmpData)

	result := new(Complex)
	err := i2s(tmpData, result)

	if err != nil {
		fmt.Printf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(expected, result) {
		fmt.Printf("results not match\nGot:\n%#v\nExpected:\n%#v", result, expected)
	}
}
