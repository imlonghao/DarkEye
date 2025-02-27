/*
Arthur: b1gCat
Email:  84500316@qq.com
 */
package plugins

/*
#include <stdlib.h>
#ifdef RDP_SUPPORT

#include <freerdp/freerdp.h>

int rdp_connect(char *server, char *port, char *domain, char *login, char *password, char *timeout) {
  int32_t err = 0;
  freerdp *instance = 0;

  instance = freerdp_new();
  if (instance == NULL || freerdp_context_new(instance) == FALSE) {
  	if (instance) freerdp_free(instance);
    return -1;
  }
  wLog *root = WLog_GetRoot();
  WLog_SetStringLogLevel(root, "OFF");

  instance->settings->Username = login;
  instance->settings->Password = password;
  instance->settings->IgnoreCertificate = TRUE;
  instance->settings->AuthenticationOnly = TRUE;
  instance->settings->ServerHostname = server;
  instance->settings->ServerPort = atoi(port);
  instance->settings->Domain = domain;
  instance->settings->TcpAckTimeout = atoi(timeout);
  freerdp_connect(instance);
  err = freerdp_get_last_error(instance->context);
  //Free
  freerdp_disconnect(instance);
  freerdp_free(instance);
  return err;
}

#else
int rdp_connect(char *server, char *port, char *domain, char *login, char *password, char *timeout) {
	return -1;
}
#endif
*/
import "C"
import (
	"context"
	"fmt"
	"github.com/zsdevX/DarkEye/superscan/dic"
	"unsafe"
)

func rdpCheck(s *Service) {
	s.crack()
}

func RdpConn(_ context.Context, s *Service, user, pass string) int {
	username := C.CString(user)
	password := C.CString(pass)
	server := C.CString(s.parent.TargetIp)
	port := C.CString(s.parent.TargetPort)
	timeout := C.CString(fmt.Sprint(Config.TimeOut))
	domain := C.CString("")

	defer func() {
		C.free(unsafe.Pointer(username))
		C.free(unsafe.Pointer(password))
		C.free(unsafe.Pointer(domain))
		C.free(unsafe.Pointer(port))
		C.free(unsafe.Pointer(server))
		C.free(unsafe.Pointer(timeout))
	}()
	ret := int(C.rdp_connect(server, port, domain, username, password, timeout))
	switch ret {
	case 0:
		// login success
		return OKDone
	case 0x00020009: // login failure
	case 0x00020014: // login failure
	case 0x00020015: // login failure
	case 0x0002000d: //?
	//case 0x00020006:
	//case 0x00020008:
	//case 0x0002000c:
	//	return OKStop
	default:
		return OKTerm
	}
	return OKNext
}
func init() {
	services["rdp"] = Service{
		name:    "rdp",
		port:    "3389",
		user:    dic.DIC_USERNAME_RDP,
		pass:    dic.DIC_PASSWORD_RDP,
		check:   rdpCheck,
		connect: RdpConn,
		thread:  8,
	}
}
