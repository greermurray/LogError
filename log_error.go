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
	if LogError("Problemas al serializar la consulta para enviar el error a registrar", errorUno) {
		return
	}

	solicitud, errorDos := http.NewRequest("POST", viper.GetString("centro_errores.endpoint"), bytes.NewBuffer(cuerpoSolicitud))
	if LogError("Problemas al conectar con el sistema de registro de error", errorDos) {
		return
	}

	solicitud.Header.Add("Content-Type", "application/json; charset=UTF-8")
	solicitud.Header.Add("Authorization", fmt.Sprintf("Bearer %s", viper.GetString("centro_errores.token")))

	respuesta, errorTres := clienteHttp.Do(solicitud)
	if LogError("Problemas con la solicitud al enviar el error a registrar", errorTres) {
		return
	}

	defer func(Body io.ReadCloser) {
		errBody := Body.Close()
		if LogError("Problemas al cerrar el cuerpo de la respuesta del error", errBody) {
			return
		}
	}(respuesta.Body)

	_, errorCuatro := io.ReadAll(respuesta.Body)
	if LogError("Problemas al leer el cuerpo de la solicitud del error", errorCuatro) {
		return
	}
}

type Opciones struct {
	RegistrarEnLaNube  bool
	EnviarNotificacion bool
}

func LogError(mensaje string, error error, opciones ...Opciones) bool {
	if error != nil {
		opcionesPorDefecto := Opciones{
			RegistrarEnLaNube:  false,
			EnviarNotificacion: false,
		}

		var opcion Opciones

		if len(opciones) > 0 {
			opcion = opciones[0]
		} else {
			opcion = opcionesPorDefecto
		}

		aplicacion := viper.GetString("aplicacion")

		_, fichero, linea, _ := runtime.Caller(1)

		var cantidadDivisiones int
		ficheroFinal := strings.Split(fichero, "/")

		if cantidad := len(ficheroFinal); cantidad > 1 {
			cantidadDivisiones = cantidad - 1
		} else {
			cantidadDivisiones = cantidad
		}

		go func() {
			if viper.GetString("centro_errores.ambiente") == "desarrollo" {
				fmt.Println(fmt.Sprintf("[FICHERO/LÍNEA]: (%s:%d) [MENSAJE]: %s [ERROR]: %s", ficheroFinal[cantidadDivisiones], linea, mensaje, error))
			}
			log.Println(fmt.Sprintf("[FICHERO/LÍNEA]: (%s:%d) [MENSAJE]: %s [ERROR]: %s", ficheroFinal[cantidadDivisiones], linea, mensaje, error))
		}()

		if opcion.RegistrarEnLaNube == true {
			go registrarError(aplicacion, fmt.Sprintf("%s:%d", ficheroFinal[cantidadDivisiones], linea), mensaje, error.Error())
		}

		if opcion.EnviarNotificacion == true {
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

	credenciales := smtp.PlainAuth("", viper.GetString("notificacion.user"), viper.GetString("notificacion.password"), viper.GetString("notificacion.host"))

	errorAlEnviarCorreo := smtp.SendMail(
		fmt.Sprintf("%s:%s",
			viper.GetString("notificacion.host"),
			viper.GetString("notificacion.port"),
		),
		credenciales,
		viper.GetString("notificacion.mail"),
		viper.GetStringSlice("notificacion.to"),
		[]byte(mensajeFormato),
	)

	if LogError("Problemas al enviar el error por correo", errorAlEnviarCorreo) {
		return
	}
}
