# README
Paquete para notificar por correo y registrar los errores en el centro de errores: https://apps.rodelag.com/admin/content/errores

Estructura que debe tener el archivo de configuración yaml de cada proyecto que use el paquete:

```console
aplicacion: ***
notificacion:
    host: ***
    port: ***
    mail: ***
    user: ***
    password: ***
    to: ***
    asunto: ***
centro_errores:
    ambiente: desarrollo
    endpoint: ***
    token: ***
```

- ##### Si la opción de **"ambiente"** está en: `desarrollo` todos los errores se imprimirán en pantalla.

### Forma de usar la librería sin manejar el error:

```console
    opcionesDeError := logerror.Opciones{
        EnviarNotificacion: true,
        RegistrarEnLaNube:  true,
    }
    logerror.LogError("Prueba con librería de registro de errores", errors.New("prueba de error"), opcionesDeError)
```

### Forma de usar la librería manejando el error:

```console
    opcionesDeError := logerror.Opciones{
        EnviarNotificacion: true,
        RegistrarEnLaNube:  true,
    }
    if logerror.LogError("Prueba con librería de registro de errores", errors.New("prueba de error"), opcionesDeError) {
        //TODO: Hacer algo si pasa algo...
    }
```

### Enviar una notificación de error:

```console
    logerror.Notificacion("Prueba con librería de registro de errores", errors.New("prueba de error"))
```