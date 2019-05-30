// Copyright (c) 2018-2019, Manticore Software LTD (http://manticoresearch.com)
// All rights reserved
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License. You should have
// received a copy of the GPL license along with this program; if you
// did not, you can find it at http://www.gnu.org/

#include "sphinxudf.h"

#include <stdio.h>
#include <string.h>
#include <inttypes.h>

#ifdef _MSC_VER
#pragma comment(linker, "/defaultlib:libmysql.lib")
#include <winsock.h>
#define snprintf _snprintf
#define DLLEXPORT __declspec(dllexport)
#else
#define DLLEXPORT
#endif

#include <mysql/mysql.h>

MYSQL * g_pMysqlConn = NULL;

char * g_szHost		= "localhost";
char * g_szUser		= "test";
char * g_szPwd		= "";
char * g_szDb		= "test";
unsigned int g_uPort = 0;

char * g_szQuery	= "SELECT title FROM example WHERE id=%"PRIi64;

/// UDF version control
/// gets called once when the library is loaded
DLLEXPORT int fetch_ver ()
{
	return SPH_UDF_VERSION;
}

/// UDF initialization
/// gets called on every query, when query begins
/// args are filled with values for a particular query
DLLEXPORT int fetch_init ( SPH_UDF_INIT * init, SPH_UDF_ARGS * args, char * error_message )
{
	if ( !g_pMysqlConn )
	{
		snprintf ( error_message, SPH_UDF_ERROR_LEN, "could not connect to MySql" );
		return 1;
	}

	if ( args->arg_count!=1 )
	{
		snprintf ( error_message, SPH_UDF_ERROR_LEN, "FETCH() requires 1 argument" );
		return 1;
	}

	return 0;
}


/// UDF deinitialization
/// gets called on every query, when query ends
DLLEXPORT void fetch_deinit ( SPH_UDF_INIT * init )
{
}

/// UDF implementation
/// gets called for every row, unless optimized away
DLLEXPORT char * fetch ( SPH_UDF_INIT * init, SPH_UDF_ARGS * args, char * error_flag )
{
	MYSQL_RES * pResult;
	MYSQL_ROW tRow;
	char szQuery[256];
	char * szResult;

	if ( !g_pMysqlConn )
		return NULL;

	snprintf ( szQuery, sizeof(szQuery), g_szQuery, *(sphinx_int64_t*)args->arg_values[0] );
	if ( mysql_query ( g_pMysqlConn, szQuery ) )
	{	
		*error_flag=1;
		return NULL;
	}

	pResult = mysql_store_result ( g_pMysqlConn );
	tRow = mysql_fetch_row ( pResult );
	if ( tRow )
	{
		szResult = strdup( tRow[0] );
		mysql_free_result ( pResult );
		return szResult;
	}

	mysql_free_result ( pResult );

	return NULL;
}


#ifdef _MSC_VER
BOOL WINAPI DllMain( HINSTANCE hinstDLL, DWORD fdwReason, LPVOID lpReserved )
#else
__attribute__((constructor)) void udfload ()
#endif
{
	g_pMysqlConn = mysql_init(NULL);

	if ( !g_pMysqlConn ) 
#ifdef _MSC_VER
		return FALSE;
#else
		return;
#endif

	if ( !mysql_real_connect( g_pMysqlConn, g_szHost, g_szUser, g_szPwd, g_szDb, g_uPort, NULL, 0) ) 
	{
		mysql_close ( g_pMysqlConn );
		g_pMysqlConn = NULL;

#ifdef _MSC_VER
		return FALSE;
#else
		return;
#endif
	}

#ifdef _MSC_VER
	return TRUE;
#endif
}


#ifndef _MSC_VER
__attribute__((destructor)) void udfcleanup ()
{
	if ( g_pMysqlConn )
		mysql_close ( g_pMysqlConn );
}
#endif
