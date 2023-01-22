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

		mensajeFormato := fmt.Sprintf("[APP]: %s [FICHERO/LÍNEA]: (%s:%d) [MENSAJE]: %s [ERROR]: %s", aplicacion, ficheroFinal[cantidadDivisiones], linea, mensaje, error)

		go func() {
			log.Printf(mensajeFormato)

			go registrarError(aplicacion, fmt.Sprintf("%s:%d", ficheroFinal[cantidadDivisiones], linea), mensaje, error.Error())

			if enviarNotificacion {
				go Notificacion(mensajeFormato, error)
			}
		}()
		return true
	}
	return false
}

func Notificacion(mensaje string, error error) {
	var (
		servidor = viper.GetString("notificacion.host")
		puerto   = viper.GetString("notificacion.port")
		usuario  = viper.GetString("notificacion.user")
		password = viper.GetString("notificacion.password")
		correoDe = viper.GetString("notificacion.mail")
		correoA  = viper.GetString("notificacion.to")
		asunto   = viper.GetString("notificacion.asunto")
	)

	msg := fmt.Sprintf("From: %s \n To: %s \n Subject: %s \n\n %s", correoDe, correoA, asunto, fmt.Sprint(mensaje, error))
	err := smtp.SendMail(fmt.Sprintf("%s:%s", servidor, puerto), smtp.PlainAuth("", usuario, password, servidor), correoDe, []string{correoA}, []byte(msg))

	LogError("Problemas al enviar la notificación", false, err)
}
