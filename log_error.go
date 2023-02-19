package logerror

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"runtime"
	"strings"
)

type graphQLConsulta struct {
	Query string `json:"query"`
}

func registrarError(aplicacion, fichero, mensaje, error string) {
	consultaGraphql := `
        mutation {
            create_errores_item(data: {
                aplicacion:"%s",
                fichero:"%s",
                mensaje:"%s",
                error:"%s"
            }) { id } 
        }
	`
	consultaGraphqlFinal := fmt.Sprintf(consultaGraphql, aplicacion, fichero, mensaje, error)

	clienteHttp := &http.Client{}

	cuerpoSolicitud, errorUno := json.Marshal(graphQLConsulta{Query: consultaGraphqlFinal})
	LogError("Problemas al con los valores", false, errorUno)

	solicitud, errorDos := http.NewRequest("POST", viper.GetString("centro_errores_endpoint"), bytes.NewBuffer(cuerpoSolicitud))
	LogError("Problemas al conectar con la api", false, errorDos)

	solicitud.Header.Add("Content-Type", "application/json; charset=UTF-8")
	solicitud.Header.Add("Authorization", fmt.Sprintf("Bearer %s", viper.GetString("centro_errores_token")))

	respuesta, errorTres := clienteHttp.Do(solicitud)
	LogError("Problemas con la solicitud", false, errorTres)

	defer func(Body io.ReadCloser) {
		errBody := Body.Close()
		LogError("Problemas al cerrar el cuerpo de la respuesta", false, errBody)
	}(respuesta.Body)

	_, errorCuatro := io.ReadAll(respuesta.Body)
	LogError("Problemas con los datos", false, errorCuatro)
}

func LogError(mensaje string, enviarNotificacion bool, error error) bool {
	if error != nil {
		aplicacion := viper.GetString("aplicacion")

		_, fichero, linea, _ := runtime.Caller(1)

		var cantidadDivisiones int
		ficheroFinal := strings.Split(fichero, "/")

		if cantidad := len(ficheroFinal); cantidad > 1 {
			cantidadDivisiones = cantidad - 1
		} else {
			cantidadDivisiones = cantidad
		}

		go log.Println(fmt.Sprintf("[FICHERO/LÍNEA]: (%s:%d) [MENSAJE]: %s [ERROR]: %s", ficheroFinal[cantidadDivisiones], linea, mensaje, error))

		go registrarError(aplicacion, fmt.Sprintf("%s:%d", ficheroFinal[cantidadDivisiones], linea), mensaje, error.Error())

		if enviarNotificacion {
			mensajeFormato := fmt.Sprintf("[APP]: %s [FICHERO/LÍNEA]: (%s:%d) [MENSAJE]: %s [ERROR]: %s", aplicacion, ficheroFinal[cantidadDivisiones], linea, mensaje, error)
			go Notificacion(mensajeFormato, error)
		}

		return true
	}
	return false
}

func Notificacion(mensaje string, error error) {
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	asunto := fmt.Sprintf("Subject: %s\n", viper.GetString("notificacion.asunto"))
	mensajeFormato := fmt.Sprintf("%s%s\n%s", asunto, mime, fmt.Sprint(mensaje, error))

	auth := smtp.PlainAuth("", viper.GetString("notificacion.user"), viper.GetString("notificacion.password"), viper.GetString("notificacion.host"))

	err := smtp.SendMail(
		fmt.Sprintf("%s:%s",
			viper.GetString("notificacion.host"),
			viper.GetString("notificacion.port"),
		),
		auth,
		viper.GetString("notificacion.mail"),
		viper.GetStringSlice("notificacion.to"),
		[]byte(mensajeFormato),
	)

	LogError("Problemas al enviar la notificación", false, err)
}
